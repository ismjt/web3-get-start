import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { network } from "hardhat";

/**
 * Hardhat文档地址：
 * https://hardhat.org/docs/learn-more/using-viem
 */
describe("MyERC20", async function () {
  const { viem } = await network.connect();
  // const publicClient = await viem.getPublicClient();

  it("Should get the MyERC20 initial name when calling the name() function", async function () {
    const myerc = await viem.deployContract("MyERC20", ["MyToken","MTK",10n]);

    assert.equal("MyToken", await myerc.read.name());
  });
});
