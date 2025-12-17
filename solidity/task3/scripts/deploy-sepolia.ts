/**
 * Sepolia 测试网部署脚本
 */
import {artifacts, network} from "hardhat";
import {encodeFunctionData} from "viem";

async function main() {
    console.log("🚀 开始部署 NFTAuction 到 Sepolia 测试网...\n");

    const { viem } = await network.connect({ network: "sepolia" });
    const [deployer] = await viem.getWalletClients();
    const publicClient = await viem.getPublicClient();

    console.log("部署账户:", deployer.account.address);
    console.log(
        "账户余额:",
        await publicClient.getBalance({ address: deployer.account.address }),
        "wei\n"
    );

    // 1. 部署价格预言机 PriceOracle
    console.log("部署 PriceOracle...");
    const priceOracle = await viem.deployContract(
        "PriceOracle",
        [],
        { client: { wallet: deployer } }
    );
    console.log("   ✅ PriceOracle:", priceOracle.address, "\n");

    // 2. 部署 NFT 合约
    console.log("部署 MajtDutchNFT...");
    const nft = await viem.deployContract("MajtDutchNFT", ["MajtDutchNFT", "MDNFT"], {
        client: { wallet: deployer },
    });
    console.log("   ✅ NFT:", nft.address);
    console.log("   📛 名称: MajtDutchNFT (MDNFT)\n");

    // 3. 部署逻辑合约
    console.log("🏭 部署逻辑合约...");
    // 部署 NFTAuction 的逻辑合约 (Implementation)，工厂合约需要这个地址来克隆/创建新的拍卖代理
    const nftAuctionImpl = await viem.deployContract("NFTAuction", []);
    console.log("   ✅ NFTAuction Impl Address:", nftAuctionImpl.address);
    // 部署 AuctionFactory 逻辑合约 (Implementation)
    const factoryImpl = await viem.deployContract("AuctionFactory", []);
    console.log("   ✅ AuctionFactory Impl Address:", factoryImpl.address, "\n");

    // 4. 部署工厂合约
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


    console.log("=" .repeat(60));
    console.log("✨ 部署完成！\n");
    console.log("📝 合约地址:");
    console.log("   PriceOracle:  ", priceOracle.address);
    console.log("   NFT Contract:    ", nft.address);
    console.log("   Auction Factory:         ", factory.address);
    console.log("\n🔗 在Etherscan中访问:  ", `https://sepolia.etherscan.io/address/${factory.address}`);
    console.log("=" .repeat(60));
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error("❌ 部署失败:", error);
        process.exit(1);
    });
