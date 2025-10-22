import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("BeggingContractModule", (m) => {
  const c = m.contract("BeggingContract", [1761122000, 1771242999]);

  return { c };
});
