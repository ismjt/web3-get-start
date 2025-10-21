// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

/// @title IntegerToRoman - 将整数转换为罗马数字
contract IntegerToRoman {

    /// @notice 将整数转换为罗马数字
    /// @param num 要转换的整数，范围 1~3999
    /// @return roman 对应的罗马数字字符串
    function intToRoman(uint256 num) public pure returns (string memory roman) {
        require(num >= 1 && num <= 3999, "Number out of range");

        // 数值表
        uint256[13] memory values = [1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1];
        // 对应符号
        string[13] memory symbols = ["M","CM","D","CD","C","XC","L","XL","X","IX","V","IV","I"];

        for (uint i = 0; i < values.length; i++) {
            while (num >= values[i]) {
                roman = string(abi.encodePacked(roman, symbols[i]));
                num -= values[i];
            }
        }

        return roman;
    }

    // =======================
    // 测试用例函数
    // =======================
    function test3749() public pure returns (string memory) {
        return intToRoman(3749); // "MMMDCCXLIX"
    }

    function test58() public pure returns (string memory) {
        return intToRoman(58); // "LVIII"
    }

    function test1994() public pure returns (string memory) {
        return intToRoman(1994); // "MCMXCIV"
    }
}
