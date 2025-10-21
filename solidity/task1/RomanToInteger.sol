// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

/// @title RomanToInteger - 将罗马数字转换整数
contract RomanToInteger {

    /// @notice 将罗马数字字符串转换为整数
    /// @param s 输入的罗马数字
    /// @return result 对应的整数值
    function romanToInt(string memory s) public pure returns (uint256 result) {
        bytes memory str = bytes(s);
        uint256 len = str.length;
        result = 0;

        for (uint256 i = 0; i < len; i++) {
            uint256 value = romanCharToInt(str[i]);

            // 判断减法规则
            /**
             * 罗马数字	正常情况	            特殊情况	    结果
               VI	    V(5) + I(1)	        正常加法	    6
               IV	    I(1) 在 V(5)左边	    特殊减法	    5 - 1 = 4
               IX	    I(1) 在 X(10)左边    特殊减法	    10 - 1 = 9
               XL	    X(10) 在 L(50)左边	特殊减法	    50 - 10 = 40
             */
            if (i < len - 1 && romanCharToInt(str[i + 1]) > value) {
                result += romanCharToInt(str[i + 1]) - value; // 特殊减法
                i++; // 跳过下一个字符
            } else {
                result += value;
            }
        }
    }

    /// @notice 将单个罗马字符转换为整数
    /// @param c 罗马字符
    /// @return 对应整数
    function romanCharToInt(bytes1 c) internal pure returns (uint256) {
        if (c == "I") return 1;
        if (c == "V") return 5;
        if (c == "X") return 10;
        if (c == "L") return 50;
        if (c == "C") return 100;
        if (c == "D") return 500;
        if (c == "M") return 1000;

        revert("Invalid Roman character");
    }

    // =======================
    // 测试用例函数
    // =======================

    function testIII() public pure returns (uint256) {
        return romanToInt("III"); // 3
    }

    function testIV() public pure returns (uint256) {
        return romanToInt("IV"); // 4
    }

    function testIX() public pure returns (uint256) {
        return romanToInt("IX"); // 9
    }

    function testLVIII() public pure returns (uint256) {
        return romanToInt("LVIII"); // 58
    }

    function testMCMXCIV() public pure returns (uint256) {
        return romanToInt("MCMXCIV"); // 1994
    }
}
