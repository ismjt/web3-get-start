/**
 * 本地测试网部署脚本
 * 用于快速部署和测试完整系统
 */
import { network, artifacts } from "hardhat";
import { expect } from "chai";
import {parseEther, encodeFunctionData, parseUnits} from "viem";
import majtDutchNFT from "../ignition/modules/MajtDutchNFT.js";

async function waitWithCountdown(seconds: number) {
    for (let i = seconds; i > 0; i--) {
        process.stdout.write(`\r等待 ${i} 秒后继续...`);
        await new Promise(resolve => setTimeout(resolve, 1000));
    }
}

async function main() {
    console.log("🚀 NFT拍卖 - 合约本地部署与测试...\n");

    const { viem } = await network.connect();
    const [deployer, seller, bidder1, bidder2, bidder3] = await viem.getWalletClients();
    const publicClient = await viem.getPublicClient();

    console.log("✅ 部署账户:", deployer.account.address);
    console.log("📈 账户余额:", await publicClient.getBalance({ address: deployer.account.address }), "wei\n");

    // 1. 部署 Mock Chainlink Aggregator
    console.log("部署 Mock Chainlink Aggregator...");
    const mockAggregator = await viem.deployContract("MockV3Aggregator", [], {
        client: { wallet: deployer },
    });
    console.log("   ✅ MockV3Aggregator:", mockAggregator.address);
    console.log("   📈 ETH 价格设置为: $2000\n");

    // 2. 部署 PriceOracle
    console.log("部署 PriceOracle...");
    const priceOracle = await viem.deployContract(
        "PriceOracle",
        [],
        { client: { wallet: deployer } }
    );
    await priceOracle.write.updateAggregator([mockAggregator.address]); // 设置为本地mock地址
    console.log("   ✅ PriceOracle:", priceOracle.address, "\n");

    // 3. 部署 NFT 合约
    console.log("部署 MajtDutchNFT...");
    const nft = await viem.deployContract("MajtDutchNFT", ["MajtDutchNFT", "MDNFT"], {
        client: { wallet: deployer },
    });
    console.log("   ✅ NFT:", nft.address);
    console.log("   📛 名称: MajtDutchNFT (MDNFT)\n");

    // 4. 部署逻辑合约
    console.log("🏭 部署逻辑合约...");
    // 部署 NFTAuction 的逻辑合约 (Implementation)，工厂合约需要这个地址来克隆/创建新的拍卖代理
    const nftAuctionImpl = await viem.deployContract("NFTAuction", []);
    console.log("   ✅ NFTAuction Impl Address:", nftAuctionImpl.address);
    // 部署 AuctionFactory 逻辑合约 (Implementation)
    const factoryImpl = await viem.deployContract("AuctionFactory", []);
    console.log("   ✅ AuctionFactory Impl Address:", factoryImpl.address, "\n");

    // 5. 部署工厂合约
    console.log("🏭 部署工厂合约...");
    const factoryArtifact = await artifacts.readArtifact("AuctionFactory");
    const initData = encodeFunctionData({
        abi: factoryArtifact.abi,
        functionName: "initialize",
        args: [
            deployer.account.address, // admin
            nftAuctionImpl.address,   // _auctionImplementation
            priceOracle.address       // _priceOracle
        ]
    })
    const factoryProxy = await viem.deployContract("UUPSProxy", [
        factoryImpl.address,
        initData
    ], {
        client: { wallet: deployer },
    });
    const factory = await viem.getContractAt("AuctionFactory", factoryProxy.address);
    console.log("   ✅ AuctionFactory Address:", factory.address, "\n");
    // console.log("   💰 默认手续费率: 2.5%\n");

    // --- 验证初始化是否成功 ---
    const storedImpl = await factory.read.auctionImplementation();
    expect(storedImpl.toLowerCase()).to.equal(nftAuctionImpl.address.toLowerCase());
    console.log("✅ Factory initialized correctly via UUPSProxy \n");

    // 6. 部署测试 ERC20
    console.log("💵 部署测试 ERC20...");
    const mockToken = await viem.deployContract(
        "MockERC20",
        ["Test Token", "tToken", 18],
        { client: { wallet: deployer } }
    );
    console.log("   ✅ MockERC20:", mockToken.address);
    const bidder3Price = parseUnits("1.345", 18); // 计划用于竞拍的ERC20token数量
    await mockToken.write.mint([bidder3.account.address, bidder3Price*2n]);
    console.log("   ✅ 向Bidder3用户【"+bidder3.account.address+"】铸造转移tToken数量20个\n");

    // 配置 Token 价格 feed
    const tokenAggregator = await viem.deployContract("MockV3Aggregator", [], {
        client: { wallet: deployer },
    });
    await tokenAggregator.write.setLatestAnswer([parseUnits("50", 8)]); // 设置代币的模拟价格
    await priceOracle.write.setTokenPriceFeed(
        [mockToken.address, tokenAggregator.address],
        { account: deployer.account }
    );
    console.log("   📈 Token 价格设置为: $1\n");

    // 7. Mint NFT
    console.log("🎁 Mint 测试 NFT...");
    await nft.write.safeMint([seller.account.address, "1"]);
    await nft.write.safeMint([seller.account.address, "2"]);
    console.log("   ✅ Minted Token #1 to:", seller.account.address);
    console.log("   ✅ Minted Token #2 to:", seller.account.address, "\n");

    const nftTokenId = 0n;
    const nft0Owner = await nft.read.ownerOf([nftTokenId]);
    console.log("   ✅ NFT Token #0 owner:", nft0Owner, "\n");

    // 调用 setApprovalForAll: 授权给 factory.address，设置为 true
    const nftAsSeller = await viem.getContractAt("MajtDutchNFT", nft.address, { client: { wallet: seller } });
    //await nftAsSeller.write.approve([factory.address, 1n]);
    const approvalTxHash = await nftAsSeller.write.setApprovalForAll(
        [factory.address, true], // [operator, approved]
    );
    // 等待授权交易确认 (在 Hardhat 测试环境中，这是必须的)
    await publicClient.waitForTransactionReceipt({ hash: approvalTxHash });

    console.log("✅ NFT 授权成功。\n");

    // 8. 创建示例拍卖
    console.log("⚡ 卖家发起 - 创建示例 ETH 拍卖...");
    const factoryAsSeller = await viem.getContractAt("AuctionFactory", factoryProxy.address, { client: { wallet: seller } });
    const createTx = await factoryAsSeller.write.createAuction(
        [
            nft.address,
            nftTokenId, // 与下方的account相匹配
            parseUnits("2", 6), // 2 参考USD，精度6
            BigInt(Math.floor(new Date().getTime() / 1000)),
            20n, // 10 second
        ]
    );

    //await waitWithCountdown(50); // 延迟秒，确保后续可在规定时间内参与拍卖

    const receipt = await publicClient.waitForTransactionReceipt({
        hash: createTx,
    });
    const auctionAddress = await factory.read.getAuction([0n]);

    console.log("   ✅ 第一个拍卖已创建:", auctionAddress);
    console.log("   🏷️  NFT Token ID: ", nftTokenId);
    console.log("   💎 起拍价:2 USD");
    console.log("   ⏰ 持续时间: 2 min\n");

    const auctionArtifact = await artifacts.readArtifact("NFTAuction");

    // 9. 模拟出价
    console.log("🎯 模拟出价...");
    const bidderAuction = await viem.getContractAt("NFTAuction", auctionAddress);
    await bidder1.writeContract({
        abi: auctionArtifact.abi,
        address: bidderAuction.address,
        functionName: "bidEth",
        args: [],
        value: parseEther("0.001") // 发送 0.001 ETH
    });
    console.log("✅ Bidder1 【"+bidder1.account.address+"】出价: 0.001 ETH\n");

    await bidder2.writeContract({
        abi: auctionArtifact.abi,
        address: bidderAuction.address,
        functionName: "bidEth",
        args: [],
        value: parseEther("0.015") // 发送 0.015 ETH
    });
    console.log("✅ Bidder2 【"+bidder2.account.address+"】出价: 0.015 ETH\n");

    console.log("🎯Bidder3检查ERC20代币余额与授权情况 ...");
    const erc20AsBidder3 = await viem.getContractAt("MockERC20", mockToken.address, { client: { wallet: bidder3 } });
    const bidder3Balance = await mockToken.read.balanceOf([bidder3.account.address]);
    console.log("Bidder3 ERC20 余额：", bidder3Balance);
    const currentAllowance = await erc20AsBidder3.read.allowance([
        bidder3.account.address,
        bidderAuction.address
    ]);
    console.log("用户【"+bidder3.account.address+"】给拍卖合约【"+bidderAuction.address+"】的ERC20授权额度:", currentAllowance.toString());
    if(currentAllowance<bidder3Price){
        console.log("发起ERC20授权...");
        const approveTx = await erc20AsBidder3.write.approve([
            bidderAuction.address,
            bidder3Price*10n,
        ]);
        console.log("ERC20授权交易已发送，等待确认...");
        await publicClient.waitForTransactionReceipt({
            hash: approveTx,
        });
        // 再次查看授权情况
        const newAllowance = await erc20AsBidder3.read.allowance([
            bidder3.account.address,
            bidderAuction.address
        ]);
        console.log("新的授权额度:", newAllowance.toString() , "tToken");
        console.log("✅ ERC20 授权完成  \n");
    }

    await bidder3.writeContract({
        abi: auctionArtifact.abi,
        address: bidderAuction.address,
        functionName: "bidERC20",
        args: [bidder3Price, mockToken.address],
    });
    console.log("✅ Bidder3 【"+bidder3.account.address+"】MockERC20 出价: 1 tToken，价值67.25 USD\n");


    const info = await bidderAuction.read.getAuctionInfo();
    console.log("📊 当前拍卖状态:");
    console.log("   最高出价情况:", info[1], "\n");

    // 等待拍卖结束
    const timeRemaining = await bidderAuction.read.timeRemaining();
    await waitWithCountdown(Number(timeRemaining));

    const canEndFlag = await bidderAuction.read.canEnd();
    if(canEndFlag){
        await bidderAuction.write.endAuction();

        console.log("   💎 验证资产转移情况...");
        const newOwner = await nft.read.ownerOf([nftTokenId]);
        console.log("New NFT Owner: ", newOwner);
        const newSellerBalance = await mockToken.read.balanceOf([seller.account.address]);
        console.log("Seller Current ERC20 Balance: ", newSellerBalance);
    }

    // 竞拍获胜者提现、转移NFT资产、转移ERC20资产
    // await bidderAuction.write.claim();

    // 10. 打印部署摘要
    console.log("\n==============");
    console.log("✨ 部署与测试完成！\n");
    console.log("📝 合约地址汇总:");
    console.log("   MockAggregator:  ", mockAggregator.address);
    console.log("   PriceOracle:  ", priceOracle.address);
    console.log("   NFT Contract:    ", nft.address);
    console.log("   Factory:         ", factory.address);
    console.log("   MockERC20:       ", mockToken.address);
    console.log("   Auction #1:      ", auctionAddress);
    console.log("\n🎮 参与账户:");
    console.log("   Deployer:        ", deployer.account.address);
    console.log("   Seller:          ", seller.account.address);
    console.log("   Bidder1:         ", bidder1.account.address);
    console.log("   Bidder2:         ", bidder2.account.address);
    console.log("   Bidder3:         ", bidder3.account.address);
    console.log("==============");
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error("❌ 部署与测试过程异常:", error);
        process.exit(1);
    });
