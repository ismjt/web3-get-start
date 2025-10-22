import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("MyERC20Module", (m) => {
  const myerc = m.contract("MyERC20", ["MyToken", "MTK", 10]);

  // m.call(myerc, "name");

  return { myerc };
});
