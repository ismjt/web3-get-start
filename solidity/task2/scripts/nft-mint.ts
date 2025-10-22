import { network } from "hardhat";

const { viem } = await network.connect({
  network: "sepolia",
  chainType: "l1",
});

// const nftAddress = "0xcae83593aed07fbf6a258d4da82d9d447b1f1cfd";
const contractAddress = "0x8368229D9AB703C6CF86021530232d03924e4C76";

const c = await viem.getContractAt("MajtNFT", contractAddress);

c.write.mintNFT(["0xA298831AD84d1A92CF99d744c4C1C827EE06A49A", "ipfs://bafkreihf5mtcysdasyvl34ccetthcl4bujhkhmcjguphtv5y6dlqalvub4"]);

