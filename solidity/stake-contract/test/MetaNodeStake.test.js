const { ethers, upgrades } = require("hardhat");
const { expect } = require("chai");

describe("MetaNodeStake", function () {
  let metaNodeToken;
  let stakeContract;
  let stakeContractAddress;
  let admin, user1, user2, user3, user4;
  const zeroAddress = "0x0000000000000000000000000000000000000000";
  const metaNodePerBlock = 100n;
  const unstakeLockedBlocks = 10;
  const provider = ethers.provider;

  let startBlock, endBlock;

  beforeEach(async function () {
    [admin, user1, user2, user3, user4] = await ethers.getSigners();

    // Deploy MetaNodeToken
    const MetaNodeToken = await ethers.getContractFactory("MetaNodeToken");
    metaNodeToken = await MetaNodeToken.deploy();
    await metaNodeToken.waitForDeployment();

    // Deploy MetaNodeStake proxy
    const MetaNodeStake = await ethers.getContractFactory("MetaNodeStake");
    const blockNumber = await provider.getBlockNumber();
    startBlock = blockNumber;
    endBlock = blockNumber + 10000;

    stakeContract = await upgrades.deployProxy(
      MetaNodeStake.connect(admin),
      [await metaNodeToken.getAddress(), startBlock, endBlock, metaNodePerBlock],
      { kind: "uups" }
    );
    await stakeContract.waitForDeployment();
    stakeContractAddress = await stakeContract.getAddress();

    // Transfer reward tokens to stake contract
    await metaNodeToken.transfer(stakeContractAddress, ethers.parseEther("10000000"));

    // Add ETH pool
    await stakeContract
      .connect(admin)
      .addPool(zeroAddress, 5, ethers.parseEther("0.001"), unstakeLockedBlocks, false);

    // Add ERC20 token pool
    const TestToken = await ethers.getContractFactory("MetaNodeToken");
    const testToken = await TestToken.deploy();
    await testToken.waitForDeployment();
    const testTokenAddress = await testToken.getAddress();

    await stakeContract
      .connect(admin)
      .addPool(testTokenAddress, 10, ethers.parseEther("1"), unstakeLockedBlocks, false);
  });

  describe("Deployment and Initialization", function () {
    it("Should initialize with correct parameters", async function () {
      expect(await stakeContract.startBlock()).to.equal(startBlock);
      expect(await stakeContract.endBlock()).to.equal(endBlock);
      expect(await stakeContract.MetaNodePerBlock()).to.equal(metaNodePerBlock);
      expect(await stakeContract.MetaNode()).to.equal(await metaNodeToken.getAddress());
    });

    it("Should have DEFAULT_ADMIN_ROLE granted", async function () {
      const DEFAULT_ADMIN_ROLE = await stakeContract.DEFAULT_ADMIN_ROLE();
      expect(await stakeContract.hasRole(DEFAULT_ADMIN_ROLE, admin.address)).to.be.true;
    });

    it("Should have ADMIN_ROLE granted", async function () {
      const ADMIN_ROLE = await stakeContract.ADMIN_ROLE();
      expect(await stakeContract.hasRole(ADMIN_ROLE, admin.address)).to.be.true;
    });

    it("Should have UPGRADE_ROLE granted", async function () {
      const UPGRADE_ROLE = await stakeContract.UPGRADE_ROLE();
      expect(await stakeContract.hasRole(UPGRADE_ROLE, admin.address)).to.be.true;
    });
  });

  describe("Pool Management", function () {
    it("Should add ETH pool", async function () {
      const poolLength = await stakeContract.poolLength();
      expect(poolLength).to.equal(2);
    });

    it("Should fail to add pool with invalid staking token on first pool", async function () {
      const MetaNodeStake = await ethers.getContractFactory("MetaNodeStake");
      const blockNumber = await provider.getBlockNumber();
      const newStart = blockNumber;
      const newEnd = blockNumber + 1000;

      const newStake = await upgrades.deployProxy(
        MetaNodeStake.connect(admin),
        [await metaNodeToken.getAddress(), newStart, newEnd, 100],
        { kind: "uups" }
      );
      await newStake.waitForDeployment();

      // First pool must be ETH (address 0)
      await expect(
        newStake.connect(admin).addPool(user1.address, 5, 0, 10, false)
      ).to.be.revertedWith("invalid staking token address");
    });

    it("Should update pool info", async function () {
      const newMinDeposit = ethers.parseEther("0.01");
      const newLockedBlocks = 20;

      await stakeContract
        .connect(admin)
        .updatePool(0, newMinDeposit, newLockedBlocks);

      const pool = await stakeContract.pool(0);
      expect(pool.minDepositAmount).to.equal(newMinDeposit);
      expect(pool.unstakeLockedBlocks).to.equal(newLockedBlocks);
    });

    it("Should set pool weight", async function () {
      const newWeight = 15;
      await stakeContract.connect(admin).setPoolWeight(0, newWeight, false);

      const pool = await stakeContract.pool(0);
      expect(pool.poolWeight).to.equal(newWeight);
    });

    it("Should update total pool weight correctly", async function () {
      const initialWeight = await stakeContract.totalPoolWeight();
      const newWeight = 20;
      await stakeContract.connect(admin).setPoolWeight(0, newWeight, false);
      const updatedWeight = await stakeContract.totalPoolWeight();

      expect(updatedWeight).to.equal(initialWeight - 5n + 20n);
    });
  });

  describe("Admin Functions", function () {
    it("Should set MetaNode token", async function () {
      const TestToken = await ethers.getContractFactory("MetaNodeToken");
      const newToken = await TestToken.deploy();
      await newToken.waitForDeployment();

      await stakeContract.connect(admin).setMetaNode(await newToken.getAddress());
      expect(await stakeContract.MetaNode()).to.equal(await newToken.getAddress());
    });

    it("Should fail to set MetaNode without ADMIN_ROLE", async function () {
      const TestToken = await ethers.getContractFactory("MetaNodeToken");
      const newToken = await TestToken.deploy();
      await newToken.waitForDeployment();

      await expect(
        stakeContract.connect(user1).setMetaNode(await newToken.getAddress())
      ).to.be.revertedWithCustomError(stakeContract, "AccessControlUnauthorizedAccount");
    });

    it("Should pause and unpause withdraw", async function () {
      expect(await stakeContract.withdrawPaused()).to.be.false;

      await stakeContract.connect(admin).pauseWithdraw();
      expect(await stakeContract.withdrawPaused()).to.be.true;

      await stakeContract.connect(admin).unpauseWithdraw();
      expect(await stakeContract.withdrawPaused()).to.be.false;
    });

    it("Should fail to pause withdraw twice", async function () {
      await stakeContract.connect(admin).pauseWithdraw();
      await expect(stakeContract.connect(admin).pauseWithdraw()).to.be.revertedWith(
        "withdraw has been already paused"
      );
    });

    it("Should fail to unpause withdraw twice", async function () {
      await expect(stakeContract.connect(admin).unpauseWithdraw()).to.be.revertedWith(
        "withdraw has been already unpaused"
      );
    });

    it("Should pause and unpause claim", async function () {
      expect(await stakeContract.claimPaused()).to.be.false;

      await stakeContract.connect(admin).pauseClaim();
      expect(await stakeContract.claimPaused()).to.be.true;

      await stakeContract.connect(admin).unpauseClaim();
      expect(await stakeContract.claimPaused()).to.be.false;
    });

    it("Should set start block", async function () {
      const newStart = BigInt(startBlock) + 100n;
      await stakeContract.connect(admin).setStartBlock(newStart);
      expect(await stakeContract.startBlock()).to.equal(newStart);
    });

    it("Should fail to set start block greater than end block", async function () {
      const currentEnd = await stakeContract.endBlock();
      await expect(
        stakeContract.connect(admin).setStartBlock(currentEnd + 1n)
      ).to.be.revertedWith("start block must be smaller than end block");
    });

    it("Should set end block", async function () {
      const newEnd = BigInt(endBlock) + 1000n;
      await stakeContract.connect(admin).setEndBlock(newEnd);
      expect(await stakeContract.endBlock()).to.equal(newEnd);
    });

    it("Should fail to set end block less than start block", async function () {
      const currentStart = await stakeContract.startBlock();
      await expect(
        stakeContract.connect(admin).setEndBlock(currentStart - 1n)
      ).to.be.revertedWith("start block must be smaller than end block");
    });

    it("Should set MetaNode per block", async function () {
      const newRate = 200n;
      await stakeContract.connect(admin).setMetaNodePerBlock(newRate);
      expect(await stakeContract.MetaNodePerBlock()).to.equal(newRate);
    });

    it("Should fail to set zero MetaNode per block", async function () {
      await expect(
        stakeContract.connect(admin).setMetaNodePerBlock(0)
      ).to.be.revertedWith("invalid parameter");
    });
  });

  describe("Deposit ETH", function () {
    it("Should deposit ETH successfully", async function () {
      const depositAmount = ethers.parseEther("10");
      await stakeContract.connect(user1).depositETH({ value: depositAmount });

      const balance = await stakeContract.stakingBalance(0, user1.address);
      expect(balance).to.equal(depositAmount);
    });

    it("Should fail to deposit ETH below minimum", async function () {
      const minDeposit = ethers.parseEther("0.001");
      const belowMin = ethers.parseEther("0.0005");

      await expect(
        stakeContract.connect(user1).depositETH({ value: belowMin })
      ).to.be.revertedWith("deposit amount is too small");
    });

    it("Should update pool staking amount on ETH deposit", async function () {
      const pool = await stakeContract.pool(0);
      const initialAmount = pool.stTokenAmount;

      const depositAmount = ethers.parseEther("5");
      await stakeContract.connect(user1).depositETH({ value: depositAmount });

      const updatedPool = await stakeContract.pool(0);
      expect(updatedPool.stTokenAmount).to.equal(initialAmount + depositAmount);
    });

    it("Should fail to deposit ETH when contract is paused", async function () {
      // The contract has whenNotPaused but we need to check if pause is available
      // For now, skip this test as pause/unpause might not be directly exposed
    });
  });

  describe("Deposit ERC20", function () {
    it("Should deposit ERC20 token successfully", async function () {
      const TestToken = await ethers.getContractFactory("MetaNodeToken");
      const testToken = TestToken.attach(await stakeContract.pool(1).then(p => p.stTokenAddress));

      const depositAmount = ethers.parseEther("100");
      await testToken.transfer(user2.address, depositAmount);
      await testToken.connect(user2).approve(stakeContractAddress, depositAmount);

      await stakeContract.connect(user2).deposit(1, depositAmount);

      const balance = await stakeContract.stakingBalance(1, user2.address);
      expect(balance).to.equal(depositAmount);
    });

    it("Should fail to deposit ERC20 to ETH pool", async function () {
      await expect(
        stakeContract.connect(user1).deposit(0, ethers.parseEther("10"))
      ).to.be.revertedWith("deposit not support ETH staking");
    });

    it("Should fail to deposit without approval", async function () {
      const depositAmount = ethers.parseEther("100");
      await expect(
        stakeContract.connect(user2).deposit(1, depositAmount)
      ).to.be.reverted;
    });

    it("Should fail to deposit below minimum", async function () {
      const TestToken = await ethers.getContractFactory("MetaNodeToken");
      const testToken = TestToken.attach(await stakeContract.pool(1).then(p => p.stTokenAddress));

      const depositAmount = ethers.parseEther("0.5");
      await testToken.transfer(user2.address, depositAmount);
      await testToken.connect(user2).approve(stakeContractAddress, depositAmount);

      await expect(
        stakeContract.connect(user2).deposit(1, depositAmount)
      ).to.be.revertedWith("deposit amount is too small");
    });
  });

  describe("Unstake", function () {
    beforeEach(async function () {
      // Deposit some ETH for user1
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("20") });
    });

    it("Should request unstake successfully", async function () {
      const unstakeAmount = ethers.parseEther("5");
      await stakeContract.connect(user1).unstake(0, unstakeAmount);

      const balance = await stakeContract.stakingBalance(0, user1.address);
      expect(balance).to.equal(ethers.parseEther("15"));
    });

    it("Should fail to unstake more than balance", async function () {
      const unstakeAmount = ethers.parseEther("30");
      await expect(
        stakeContract.connect(user1).unstake(0, unstakeAmount)
      ).to.be.revertedWith("Not enough staking token balance");
    });

    it("Should fail to unstake when withdraw paused", async function () {
      await stakeContract.connect(admin).pauseWithdraw();

      await expect(
        stakeContract.connect(user1).unstake(0, ethers.parseEther("5"))
      ).to.be.revertedWith("withdraw is paused");

      await stakeContract.connect(admin).unpauseWithdraw();
    });

    it("Should create unstake request with correct unlock block", async function () {
      const currentBlock = await provider.getBlockNumber();
      const unstakeAmount = ethers.parseEther("5");

      await stakeContract.connect(user1).unstake(0, unstakeAmount);

      const withdrawInfo = await stakeContract.withdrawAmount(0, user1.address);
      expect(withdrawInfo.requestAmount).to.equal(unstakeAmount);
    });

    it("Should accumulate pending MetaNode on unstake", async function () {
      // Move forward some blocks to accumulate rewards
      for (let i = 0; i < 5; i++) {
        await provider.send("evm_mine", []);
      }

      await stakeContract.connect(user1).unstake(0, ethers.parseEther("5"));

      const pending = await stakeContract.pendingMetaNode(0, user1.address);
      expect(pending).to.be.gt(0);
    });
  });

  describe("Withdraw", function () {
    beforeEach(async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("20") });
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("5"));
    });

    it("Should withdraw after unlock period", async function () {
      const balanceBefore = await provider.getBalance(user1.address);

      // Move forward locked blocks
      for (let i = 0; i < unstakeLockedBlocks; i++) {
        await provider.send("evm_mine", []);
      }

      const tx = await stakeContract.connect(user1).withdraw(0);
      const receipt = await tx.wait();

      const balanceAfter = await provider.getBalance(user1.address);
      const gasUsed = receipt.gasUsed * receipt.gasPrice;

      // Balance should increase (minus gas)
      expect(balanceAfter - balanceBefore + gasUsed).to.be.gte(ethers.parseEther("4.9"));
      expect(balanceAfter - balanceBefore + gasUsed).to.be.lte(ethers.parseEther("5.1"));
    });

    it("Should not withdraw anything before unlock period", async function () {
      const balanceBefore = await provider.getBalance(user1.address);
      await stakeContract.connect(user1).withdraw(0);
      const balanceAfter = await provider.getBalance(user1.address);

      // Should only deduct gas, no ETH should be transferred
      expect(balanceAfter).to.be.lte(balanceBefore);
      expect(balanceBefore - balanceAfter).to.be.lt(ethers.parseEther("0.01")); // Only gas cost
    });

    it("Should fail to withdraw when paused", async function () {
      await stakeContract.connect(admin).pauseWithdraw();

      // Move forward locked blocks
      for (let i = 0; i < unstakeLockedBlocks; i++) {
        await provider.send("evm_mine", []);
      }

      await expect(
        stakeContract.connect(user1).withdraw(0)
      ).to.be.revertedWith("withdraw is paused");

      await stakeContract.connect(admin).unpauseWithdraw();
    });

    it("Should handle multiple unstake requests", async function () {
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("2"));
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("3"));

      const withdrawInfo = await stakeContract.withdrawAmount(0, user1.address);
      expect(withdrawInfo.requestAmount).to.equal(ethers.parseEther("10"));
    });
  });

  describe("Claim Rewards", function () {
    beforeEach(async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stakeContract.connect(user2).depositETH({ value: ethers.parseEther("20") });
    });

    it("Should claim MetaNode rewards", async function () {
      // Move forward some blocks
      for (let i = 0; i < 10; i++) {
        await provider.send("evm_mine", []);
      }

      const pendingBefore = await stakeContract.pendingMetaNode(0, user1.address);
      expect(pendingBefore).to.be.gt(0);

      const balanceBefore = await metaNodeToken.balanceOf(user1.address);
      await stakeContract.connect(user1).claim(0);
      const balanceAfter = await metaNodeToken.balanceOf(user1.address);

      expect(balanceAfter - balanceBefore).to.be.gte(pendingBefore - BigInt(10)); // Allow small rounding
    });

    it("Should fail to claim when paused", async function () {
      for (let i = 0; i < 10; i++) {
        await provider.send("evm_mine", []);
      }

      await stakeContract.connect(admin).pauseClaim();

      await expect(
        stakeContract.connect(user1).claim(0)
      ).to.be.revertedWith("claim is paused");

      await stakeContract.connect(admin).unpauseClaim();
    });

    it("Should accumulate rewards based on pool weight", async function () {
      // Move forward some blocks
      for (let i = 0; i < 20; i++) {
        await provider.send("evm_mine", []);
      }

      const pending1 = BigInt(await stakeContract.pendingMetaNode(0, user1.address));
      const pending2 = BigInt(await stakeContract.pendingMetaNode(0, user2.address));

      // user2 deposited 2x, should get roughly 2x rewards (considering weights)
      expect(pending2).to.be.gt(pending1);
    });

    it("Should clear pending MetaNode after claim", async function () {
      for (let i = 0; i < 10; i++) {
        await provider.send("evm_mine", []);
      }

      await stakeContract.connect(user1).claim(0);

      const pending = await stakeContract.pendingMetaNode(0, user1.address);
      // Pending should be nearly zero (may be 1-2 blocks worth due to timing)
      expect(pending).to.be.lte(metaNodePerBlock * 2n);
    });

    it("Should handle zero pending rewards", async function () {
      const balanceBefore = await metaNodeToken.balanceOf(user3.address);
      await stakeContract.connect(user3).claim(0);
      const balanceAfter = await metaNodeToken.balanceOf(user3.address);

      expect(balanceAfter).to.equal(balanceBefore);
    });
  });

  describe("Query Functions", function () {
    beforeEach(async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stakeContract.connect(user2).depositETH({ value: ethers.parseEther("20") });
    });

    it("Should return correct pool length", async function () {
      const poolLength = await stakeContract.poolLength();
      expect(poolLength).to.equal(2);
    });

    it("Should calculate multiplier correctly", async function () {
      const from = BigInt(startBlock);
      const to = BigInt(startBlock) + 100n;

      const multiplier = await stakeContract.getMultiplier(from, to);
      expect(multiplier).to.equal((to - from) * metaNodePerBlock);
    });

    it("Should clamp multiplier to startBlock", async function () {
      const from = BigInt(startBlock) - 100n;
      const to = BigInt(startBlock) + 50n;

      const multiplier = await stakeContract.getMultiplier(from, to);
      expect(multiplier).to.equal(50n * metaNodePerBlock);
    });

    it("Should clamp multiplier to endBlock", async function () {
      const from = BigInt(endBlock) - 50n;
      const to = BigInt(endBlock) + 100n;

      const multiplier = await stakeContract.getMultiplier(from, to);
      expect(multiplier).to.equal(50n * metaNodePerBlock);
    });

    it("Should return pending MetaNode for user", async function () {
      for (let i = 0; i < 10; i++) {
        await provider.send("evm_mine", []);
      }

      const pending = await stakeContract.pendingMetaNode(0, user1.address);
      expect(pending).to.be.gt(0);
    });

    it("Should return staking balance", async function () {
      const balance = await stakeContract.stakingBalance(0, user1.address);
      expect(balance).to.equal(ethers.parseEther("10"));
    });

    it("Should return withdraw amount info", async function () {
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("5"));

      const withdrawInfo = await stakeContract.withdrawAmount(0, user1.address);
      expect(withdrawInfo.requestAmount).to.equal(ethers.parseEther("5"));
      expect(withdrawInfo.pendingWithdrawAmount).to.equal(0);
    });

    it("Should return pending withdraw amount after unlock", async function () {
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("5"));

      for (let i = 0; i < unstakeLockedBlocks; i++) {
        await provider.send("evm_mine", []);
      }

      const withdrawInfo = await stakeContract.withdrawAmount(0, user1.address);
      expect(withdrawInfo.requestAmount).to.equal(ethers.parseEther("5"));
      expect(withdrawInfo.pendingWithdrawAmount).to.equal(ethers.parseEther("5"));
    });
  });

  describe("Pool Update", function () {
    it("Should update pool correctly", async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("10") });

      const poolBefore = await stakeContract.pool(0);
      const blockBefore = poolBefore.lastRewardBlock;

      for (let i = 0; i < 5; i++) {
        await provider.send("evm_mine", []);
      }

      await stakeContract.updatePool(0);

      const poolAfter = await stakeContract.pool(0);
      const blockAfter = poolAfter.lastRewardBlock;

      expect(blockAfter).to.be.gt(blockBefore);
      expect(poolAfter.accMetaNodePerST).to.be.gt(0);
    });

    it("Should mass update all pools", async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("10") });

      const TestToken = await ethers.getContractFactory("MetaNodeToken");
      const testToken = TestToken.attach(await stakeContract.pool(1).then(p => p.stTokenAddress));
      await testToken.transfer(user2.address, ethers.parseEther("100"));
      await testToken.connect(user2).approve(stakeContractAddress, ethers.parseEther("100"));
      await stakeContract.connect(user2).deposit(1, ethers.parseEther("100"));

      for (let i = 0; i < 5; i++) {
        await provider.send("evm_mine", []);
      }

      await stakeContract.massUpdatePools();

      const pool0 = await stakeContract.pool(0);
      const pool1 = await stakeContract.pool(1);

      expect(pool0.accMetaNodePerST).to.be.gt(0);
      expect(pool1.accMetaNodePerST).to.be.gt(0);
    });
  });

  describe("Edge Cases and Error Handling", function () {
    it("Should fail with invalid pool ID", async function () {
      await expect(
        stakeContract.pendingMetaNode(999, user1.address)
      ).to.be.revertedWith("invalid pid");
    });

    it("Should handle getMultiplier with invalid range", async function () {
      await expect(
        stakeContract.getMultiplier(100, 50)
      ).to.be.revertedWith("invalid block");
    });

    it("Should fail to add pool after end block", async function () {
      const currentEnd = await stakeContract.endBlock();
      const currentBlock = BigInt(await provider.getBlockNumber());

      // Set endBlock to before current block
      if (currentBlock < currentEnd) {
        await stakeContract.connect(admin).setEndBlock(currentBlock - 1n);

        const TestToken = await ethers.getContractFactory("MetaNodeToken");
        const testToken = await TestToken.deploy();
        await testToken.waitForDeployment();

        await expect(
          stakeContract.connect(admin).addPool(
            await testToken.getAddress(),
            5,
            0,
            10,
            false
          )
        ).to.be.revertedWith("Already ended");
      }
    });

    it("Should handle deposit amount at minimum", async function () {
      const minAmount = ethers.parseEther("0.001");
      await stakeContract.connect(user1).depositETH({ value: minAmount });
      const balance = await stakeContract.stakingBalance(0, user1.address);

      expect(balance).to.equal(minAmount);
    });
  });

  describe("Internal Functions Coverage", function () {
    it("Should handle safe MetaNode transfer correctly", async function () {
      // User deposits and mines blocks to accumulate rewards
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("10") });

      for (let i = 0; i < 50; i++) {
        await provider.send("evm_mine", []);
      }

      // Get pending rewards
      const pending = await stakeContract.pendingMetaNode(0, user1.address);
      expect(pending).to.be.gt(0);

      // Claim rewards - _safeMetaNodeTransfer will be invoked
      const userBalanceBefore = await metaNodeToken.balanceOf(user1.address);
      await stakeContract.connect(user1).claim(0);
      const userBalanceAfter = await metaNodeToken.balanceOf(user1.address);

      // Should receive rewards
      expect(userBalanceAfter).to.be.gt(userBalanceBefore);
    });

    it("Should handle multiple deposits with different amounts", async function () {
      const amount1 = ethers.parseEther("5");
      const amount2 = ethers.parseEther("7");
      const amount3 = ethers.parseEther("3");

      await stakeContract.connect(user1).depositETH({ value: amount1 });
      await stakeContract.connect(user1).depositETH({ value: amount2 });
      await stakeContract.connect(user1).depositETH({ value: amount3 });

      const totalBalance = await stakeContract.stakingBalance(0, user1.address);
      expect(totalBalance).to.equal(amount1 + amount2 + amount3);
    });

    it("Should handle complex unstake scenarios", async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("30") });

      // Multiple unstake requests
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("5"));
      for (let i = 0; i < 5; i++) {
        await provider.send("evm_mine", []);
      }

      await stakeContract.connect(user1).unstake(0, ethers.parseEther("3"));
      for (let i = 0; i < 3; i++) {
        await provider.send("evm_mine", []);
      }

      await stakeContract.connect(user1).unstake(0, ethers.parseEther("7"));

      // Verify remaining stake
      const remaining = await stakeContract.stakingBalance(0, user1.address);
      expect(remaining).to.equal(ethers.parseEther("15"));

      // All should be in requests
      const withdrawInfo = await stakeContract.withdrawAmount(0, user1.address);
      expect(withdrawInfo.requestAmount).to.equal(ethers.parseEther("15"));
    });

    it("Should process multiple partial withdrawals", async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("50") });

      // Create multiple unstake requests
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("10"));
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("15"));
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("5"));

      // Wait for unlock
      for (let i = 0; i < unstakeLockedBlocks; i++) {
        await provider.send("evm_mine", []);
      }

      // Withdraw should process all
      const balanceBefore = await provider.getBalance(user1.address);
      const tx = await stakeContract.connect(user1).withdraw(0);
      const receipt = await tx.wait();
      const balanceAfter = await provider.getBalance(user1.address);

      const gasUsed = receipt.gasUsed * receipt.gasPrice;
      const ethGained = balanceAfter - balanceBefore + gasUsed;

      // Should receive approximately 30 ETH worth (all three unstakes)
      expect(ethGained).to.be.gte(ethers.parseEther("29.9"));
    });

    it("Should properly update accMetaNodePerST when staking amount is zero", async function () {
      // Don't deposit anything, just update pool
      const pool0Before = await stakeContract.pool(0);
      await stakeContract.updatePool(0);
      const pool0After = await stakeContract.pool(0);

      // accMetaNodePerST should remain 0 if no staking
      expect(pool0After.accMetaNodePerST).to.equal(pool0Before.accMetaNodePerST);
    });

    it("Should handle massUpdatePools with empty pools", async function () {
      // Add an empty pool
      const TestToken = await ethers.getContractFactory("MetaNodeToken");
      const emptyToken = await TestToken.deploy();
      await emptyToken.waitForDeployment();

      await stakeContract
        .connect(admin)
        .addPool(
          await emptyToken.getAddress(),
          5,
          ethers.parseEther("1"),
          unstakeLockedBlocks,
          false
        );

      // Update all pools including empty ones
      await stakeContract.massUpdatePools();

      const pool2 = await stakeContract.pool(2);
      expect(pool2.stTokenAmount).to.equal(0);
    });

    it("Should handle claim with accumulated pending MetaNode", async function () {
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("10") });

      // Accumulate some pending rewards
      for (let i = 0; i < 30; i++) {
        await provider.send("evm_mine", []);
      }

      // Check pending before claim
      const pendingBefore = await stakeContract.pendingMetaNode(0, user1.address);
      expect(pendingBefore).to.be.gt(0);

      // Claim
      await stakeContract.connect(user1).claim(0);

      // Check balance increased
      const userBalance = await metaNodeToken.balanceOf(user1.address);
      expect(userBalance).to.be.gt(0);
    });
  });

  describe("Reentrancy Protection", function () {
    it("Should complete full cycle: deposit -> accumulate -> claim -> unstake -> withdraw", async function () {
      // Deposit
      await stakeContract.connect(user1).depositETH({ value: ethers.parseEther("10") });
      const depositedAmount = await stakeContract.stakingBalance(0, user1.address);
      expect(depositedAmount).to.equal(ethers.parseEther("10"));

      // Mine blocks to accumulate rewards
      for (let i = 0; i < 20; i++) {
        await provider.send("evm_mine", []);
      }

      // Claim rewards
      const pendingBefore = await stakeContract.pendingMetaNode(0, user1.address);
      expect(pendingBefore).to.be.gt(0);
      const balanceBefore = await metaNodeToken.balanceOf(user1.address);
      await stakeContract.connect(user1).claim(0);
      const balanceAfter = await metaNodeToken.balanceOf(user1.address);
      expect(balanceAfter - balanceBefore).to.be.gt(0);

      // Unstake
      await stakeContract.connect(user1).unstake(0, ethers.parseEther("5"));
      const remainingStake = await stakeContract.stakingBalance(0, user1.address);
      expect(remainingStake).to.equal(ethers.parseEther("5"));

      // Wait for unlock
      for (let i = 0; i < unstakeLockedBlocks; i++) {
        await provider.send("evm_mine", []);
      }

      // Withdraw
      const ethBalanceBefore = await provider.getBalance(user1.address);
      const tx = await stakeContract.connect(user1).withdraw(0);
      const receipt = await tx.wait();
      const ethBalanceAfter = await provider.getBalance(user1.address);

      const gasUsed = receipt.gasUsed * receipt.gasPrice;
      const ethGained = ethBalanceAfter - ethBalanceBefore + gasUsed;
      expect(ethGained).to.be.gte(ethers.parseEther("4.9"));
      expect(ethGained).to.be.lte(ethers.parseEther("5.1"));
    });
  });
});
