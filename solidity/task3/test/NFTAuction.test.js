import { network } from "hardhat";
import { describe, it, before } from "node:test";
import { parseEther, parseUnits, zeroAddress } from "viem";
import upgrades from "@openzeppelin/hardhat-upgrades";

describe("NFTAuction", async function () {
    const { viem } = await network.connect();

    let accounts;
    let seller, bidder1, bidder2, factory;
    let publicClient;

    let nftAddress;
    let erc20Address;
    let ethAggregatorAddress;
    let usdcAggregatorAddress;
    let auctionAddress;

    const tokenId = 1n;
    const startPriceUSD = parseUnits("1000", 18); // $1000
    const duration = 60n * 60n; // 1 hour

    before(async () => {
        const walletClients = await viem.getWalletClients();
        accounts = walletClients.map(w => w.account);
        [seller, bidder1, bidder2, factory] = accounts;
        publicClient = await viem.getPublicClient();
    });

    beforeEach(async () => {
        // Deploy TestERC721
        const nft = await viem.deployContract("TestERC721", []);
        nftAddress = nft.address;
        await nft.write.mint([seller.address, tokenId]);

        // Deploy TestERC20
        const erc20 = await viem.deployContract("TestERC20", ["TestToken", "TST"]);
        erc20Address = erc20.address;
        await erc20.write.mint([bidder1.address, parseUnits("10000", 18)]);
        await erc20.write.mint([bidder2.address, parseUnits("10000", 18)]);

        // Deploy MockV3Aggregators
        const MockV3 = await viem.getContractFactory("MockV3Aggregator");
        const ethAgg = await MockV3.deploy([8, 2000e8]); // 8 decimals, $2000
        const usdcAgg = await MockV3.deploy([8, 1e8]);   // $1
        ethAggregatorAddress = ethAgg.address;
        usdcAggregatorAddress = usdcAgg.address;

        // Deploy NFTAuction proxy (UUPS)
        const Auction = await viem.getContractFactory("NFTAuction");
        const auctionImpl = await Auction.deploy(); // deploy implementation
        const AuctionProxy = await viem.getContractFactory("NFTAuction"); // for upgrades
        // ⚠️ OpenZeppelin upgrades still uses ethers internally, so we use hybrid approach
        const auctionEthers = await upgrades.deployProxy(AuctionProxy, [], {
            initializer: false,
            kind: "uups",
        });
        auctionAddress = await auctionEthers.getAddress();

        // Approve NFT
        const nftViem = await viem.getContractAt("TestERC721", nftAddress);
        await nftViem.write.approve([auctionAddress, tokenId], { account: seller });
    });

    describe("ETH Payment Auction", () => {
        let auction;
        let startTime;

        beforeEach(async () => {
            auction = await viem.getContractAt("NFTAuction", auctionAddress);
            const block = await publicClient.getBlock();
            startTime = block.timestamp + 10n;

            await auction.write.initialize([
                seller.address,
                nftAddress,
                tokenId,
                startPriceUSD,
                startTime,
                duration,
                zeroAddress,
                ethAggregatorAddress,
                factory.address
            ], { account: seller });

            // Mine block at startTime + 1
            await network.provider.send("evm_setNextBlockTimestamp", [Number(startTime + 1n)]);
            await network.provider.send("evm_mine", []);
        });

        it("Should allow bidding with ETH", async () => {
            const minBid = parseEther("0.5"); // $1000 / $2000 = 0.5 ETH
            const txHash = await auction.write.bid([0n], {
                account: bidder1,
                value: minBid
            });

            const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(receipt.status).to.equal("success");

            const usdValue = await auction.read.calculateUsdValue([minBid]);
            expect(usdValue).to.equal(startPriceUSD);
        });

        it("Should reject bid below start price", async () => {
            const lowBid = parseEther("0.4");
            await expect(
                auction.write.bid([0n], { account: bidder1, value: lowBid })
            ).to.be.rejectedWith(/Bid below start price/);
        });

        it("Should allow higher bid and refund previous bidder", async () => {
            const bid1 = parseEther("0.6");
            const bid2 = parseEther("0.7");

            await auction.write.bid([0n], { account: bidder1, value: bid1 });
            await auction.write.bid([0n], { account: bidder2, value: bid2 });

            const balance = await publicClient.getBalance({ address: auctionAddress });
            expect(balance).to.equal(bid2);

            const pending = await auction.read.pendingReturns([bidder1.address]);
            expect(pending).to.equal(bid1);
        });

        it("Should allow winner to claim NFT and seller to receive ETH", async () => {
            const bid = parseEther("0.6");
            await auction.write.bid([0n], { account: bidder1, value: bid });

            // Advance time
            await hre.network.provider.send("evm_increaseTime", [Number(duration + 10n)]);
            await hre.network.provider.send("evm_mine", []);

            await auction.write.endAuction({ account: seller });
            const sellerBalanceBefore = await publicClient.getBalance({ address: seller.address });
            const txHash = await auction.write.claim({ account: bidder1 });
            const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });

            const sellerBalanceAfter = await publicClient.getBalance({ address: seller.address });
            const gasUsed = receipt.gasUsed * receipt.effectiveGasPrice;

            // Approximate check
            expect(sellerBalanceAfter + gasUsed).to.be.approximately(sellerBalanceBefore + bid, 1e10);

            const nft = await hre.viem.getContractAt("TestERC721", nftAddress);
            const owner = await nft.read.ownerOf([tokenId]);
            expect(owner).to.equal(bidder1.address);

            const data = await auction.read.auctionData();
            expect(data[5]).to.be.true; // claimed is the 6th field (index 5)
        });
    });

    // Note: Full Viem-only upgradeable deployment is not yet supported by OpenZeppelin.
    // So we use a hybrid: ethers for proxy deployment, Viem for interaction.
});