import { network } from "hardhat";

const { viem } = await network.connect({
  network: "sepolia",
  chainType: "l1",
});

// https://sepolia.etherscan.io/address/0xe59E77d2659F199F9cb82964f7673C1f75C2baB8
const contractAddress = "0xe59E77d2659F199F9cb82964f7673C1f75C2baB8";

const c = await viem.getContractAt("BeggingContract", contractAddress);

const addDonation1 = await c.read.getDonation(["0xA298831AD84d1A92CF99d744c4C1C827EE06A49A"]);
console.log("[Before]Address Donation Amount: " + addDonation1);

await c.write.donate({value: 10_000_000_000_000_000n}); // 0.01ETH

const addDonation2 = await c.read.getDonation(["0xA298831AD84d1A92CF99d744c4C1C827EE06A49A"]);
console.log("[After]Address Donation Amount: " + addDonation2);

const balance1 = await c.read.getBalance();
console.log("[Before Withdraw] Contract Balance: " + balance1);
await c.write.withdraw();
const balance2 = await c.read.getBalance();
console.log("[After Withdraw] Contract Balance: " + balance2);


