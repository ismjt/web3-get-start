import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("MajtNFTModule", (m) => {
  const c = m.contract("MajtNFT", ["MajtNFT", "MajtNFT"]);

  // 测试钱包一
  m.call(c, "mintNFT", ["0xeCBB02Aac0ea98E218b8E3A891F04a87a2663970", "ipfs://bafkreihf5mtcysdasyvl34ccetthcl4bujhkhmcjguphtv5y6dlqalvub4"]);

  // 测试钱包二
  m.call(c, "mintNFT", ["0xA298831AD84d1A92CF99d744c4C1C827EE06A49A", "ipfs://bafkreihf5mtcysdasyvl34ccetthcl4bujhkhmcjguphtv5y6dlqalvub4"]);

  return { c };
});
