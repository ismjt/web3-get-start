// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

contract ReverseString{

    function reverse(string memory input) public pure returns (string memory) {
        bytes memory strBytes = bytes(input);
        uint256 len = strBytes.length;
        bytes memory reversed = new bytes(len);

        for (uint256 i = 0; i < len; i++) {
            reversed[i] = strBytes[len - 1 - i];
        }

        return string(reversed);
    }
}