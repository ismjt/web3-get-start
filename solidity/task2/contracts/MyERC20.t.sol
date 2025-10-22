// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import "forge-std/src/Test.sol";
import {MyERC20} from "./MyERC20.sol";

contract MyERC20Test is Test {
    MyERC20 myerc;

  function setUp() public {
      myerc = new MyERC20("MyToken","MTK",10);
  }

  function test_InitialName() public view {
    require(keccak256(bytes(myerc.name())) == keccak256(bytes("MyToken")), "Initial value should be MyToken");
  }
}
