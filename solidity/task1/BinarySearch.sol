// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title BinarySearch
 * @dev 实现二分查找算法，并包含一个内置测试用例
 */
contract BinarySearch {

    /**
     * @notice 在有序数组中使用二分查找定位目标值
     * @param arr 升序排列的数组
     * @param target 要查找的目标值
     * @return index 如果找到，返回索引；否则返回 uint256(-1)
     */
    function binarySearch(uint[] memory arr, uint target) public pure returns (int256) {
        uint left = 0;
        uint right = arr.length;

        while (left < right) {
            uint mid = left + (right - left) / 2;

            if (arr[mid] == target) {
                return int256(mid);
            } else if (arr[mid] < target) {
                left = mid + 1;
            } else {
                right = mid;
            }
        }

        return -1; // 未找到
    }

    /**
     * @notice 内置测试函数：验证二分查找的正确性
     * @return success 测试是否通过
     */
    function runTest() external pure returns (bool success) {
        // 创建测试数组 [1, 3, 5, 7, 9]
        uint[] memory arr = new uint[](5);
        arr[0] = 1;
        arr[1] = 3;
        arr[2] = 5;
        arr[3] = 7;
        arr[4] = 9;

        // 测试 1: 查找存在的元素
        int256 result1 = binarySearch(arr, 5);
        require(result1 == 2, "Test failed: should find 5 at index 2");

        // 测试 2: 查找不存在的元素
        int256 result2 = binarySearch(arr, 4);
        require(result2 == -1, "Test failed: should not find 4");

        // 测试 3: 查找第一个元素
        int256 result3 = binarySearch(arr, 1);
        require(result3 == 0, "Test failed: should find 1 at index 0");

        // 测试 4: 查找最后一个元素
        int256 result4 = binarySearch(arr, 9);
        require(result4 == 4, "Test failed: should find 9 at index 4");

        // 测试 5: 查找超出范围的值
        int256 result5 = binarySearch(arr, 10);
        require(result5 == -1, "Test failed: should not find 10");

        // 如果所有 require 通过，返回 true
        return true;
    }
}