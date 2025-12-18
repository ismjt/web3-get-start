const { ethers } = require("hardhat");
const { expect } = require("chai");

describe("MetaNodeToken", function () {
  let metaNodeToken;
  let owner, addr1, addr2;

  beforeEach(async function () {
    [owner, addr1, addr2] = await ethers.getSigners();
    const MetaNodeToken = await ethers.getContractFactory("MetaNodeToken");
    metaNodeToken = await MetaNodeToken.deploy();
    await metaNodeToken.waitForDeployment();
  });

  describe("Deployment", function () {
    it("Should have correct name and symbol", async function () {
      expect(await metaNodeToken.name()).to.equal("MetaNodeToken");
      expect(await metaNodeToken.symbol()).to.equal("MetaNode");
    });

    it("Should mint initial supply to owner", async function () {
      const initialSupply = ethers.parseEther("10000000");
      const ownerBalance = await metaNodeToken.balanceOf(owner.address);
      expect(ownerBalance).to.equal(initialSupply);
    });

    it("Should have 18 decimals", async function () {
      expect(await metaNodeToken.decimals()).to.equal(18);
    });

    it("Should have correct total supply", async function () {
      const totalSupply = await metaNodeToken.totalSupply();
      const expected = ethers.parseEther("10000000");
      expect(totalSupply).to.equal(expected);
    });
  });

  describe("Transfers", function () {
    it("Should transfer tokens between accounts", async function () {
      const transferAmount = ethers.parseEther("100");
      await metaNodeToken.transfer(addr1.address, transferAmount);
      const balance = await metaNodeToken.balanceOf(addr1.address);
      expect(balance).to.equal(transferAmount);
    });

    it("Should fail if sender doesn't have enough tokens", async function () {
      const ownerBalance = await metaNodeToken.balanceOf(owner.address);
      await expect(
        metaNodeToken.connect(addr1).transfer(owner.address, ownerBalance + BigInt(1))
      ).to.be.revertedWithCustomError(metaNodeToken, "ERC20InsufficientBalance");
    });

    it("Should update balances after transfers", async function () {
      const transferAmount = ethers.parseEther("50");
      await metaNodeToken.transfer(addr1.address, transferAmount);
      await metaNodeToken.transfer(addr2.address, transferAmount);

      const finalBalanceOwner = await metaNodeToken.balanceOf(owner.address);
      const finalBalanceAddr1 = await metaNodeToken.balanceOf(addr1.address);
      const finalBalanceAddr2 = await metaNodeToken.balanceOf(addr2.address);

      expect(finalBalanceOwner).to.equal(ethers.parseEther("10000000") - transferAmount * BigInt(2));
      expect(finalBalanceAddr1).to.equal(transferAmount);
      expect(finalBalanceAddr2).to.equal(transferAmount);
    });
  });

  describe("Approvals and TransferFrom", function () {
    it("Should approve tokens for spending", async function () {
      const approveAmount = ethers.parseEther("100");
      await metaNodeToken.approve(addr1.address, approveAmount);
      const allowance = await metaNodeToken.allowance(owner.address, addr1.address);
      expect(allowance).to.equal(approveAmount);
    });

    it("Should transferFrom approved tokens", async function () {
      const approveAmount = ethers.parseEther("100");
      await metaNodeToken.approve(addr1.address, approveAmount);
      await metaNodeToken
        .connect(addr1)
        .transferFrom(owner.address, addr2.address, approveAmount);

      const balanceAddr2 = await metaNodeToken.balanceOf(addr2.address);
      expect(balanceAddr2).to.equal(approveAmount);

      const allowance = await metaNodeToken.allowance(owner.address, addr1.address);
      expect(allowance).to.equal(0);
    });

    it("Should fail transferFrom without approval", async function () {
      const transferAmount = ethers.parseEther("100");
      await expect(
        metaNodeToken
          .connect(addr1)
          .transferFrom(owner.address, addr2.address, transferAmount)
      ).to.be.revertedWithCustomError(metaNodeToken, "ERC20InsufficientAllowance");
    });

    it("Should fail transferFrom with insufficient allowance", async function () {
      const approveAmount = ethers.parseEther("50");
      const transferAmount = ethers.parseEther("100");
      await metaNodeToken.approve(addr1.address, approveAmount);

      await expect(
        metaNodeToken
          .connect(addr1)
          .transferFrom(owner.address, addr2.address, transferAmount)
      ).to.be.revertedWithCustomError(metaNodeToken, "ERC20InsufficientAllowance");
    });

    it("Should handle multiple approvals", async function () {
      const firstAmount = ethers.parseEther("100");
      const secondAmount = ethers.parseEther("50");

      await metaNodeToken.approve(addr1.address, firstAmount);
      let allowance = await metaNodeToken.allowance(owner.address, addr1.address);
      expect(allowance).to.equal(firstAmount);

      // Approve again with different amount
      await metaNodeToken.approve(addr1.address, secondAmount);
      allowance = await metaNodeToken.allowance(owner.address, addr1.address);
      expect(allowance).to.equal(secondAmount);
    });

    it("Should handle zero approval", async function () {
      const amount = ethers.parseEther("100");
      await metaNodeToken.approve(addr1.address, amount);
      await metaNodeToken.approve(addr1.address, 0);

      const allowance = await metaNodeToken.allowance(owner.address, addr1.address);
      expect(allowance).to.equal(0);
    });
  });
});
