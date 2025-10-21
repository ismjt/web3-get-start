// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title MergeSortedArray - 合并两个有序数组
/// @notice 将两个有序数组合并成一个有序数组
contract MergeSortedArray {

    /// @notice 合并两个有序数组
    /// @param arr1 第一个有序数组
    /// @param arr2 第二个有序数组
    /// @return merged 合并后的有序数组
    function mergeSorted(uint256[] memory arr1, uint256[] memory arr2) public pure returns (uint256[] memory merged) {
        uint256 len1 = arr1.length;
        uint256 len2 = arr2.length;
        merged = new uint256[](len1 + len2);

        uint256 i = 0; // arr1 指针
        uint256 j = 0; // arr2 指针
        uint256 k = 0; // merged 指针

        while (i < len1 && j < len2) {
            if (arr1[i] <= arr2[j]) {
                merged[k] = arr1[i];
                i++;
            } else {
                merged[k] = arr2[j];
                j++;
            }
            k++;
        }

        // 剩余元素复制
        while (i < len1) {
            merged[k] = arr1[i];
            i++;
            k++;
        }

        while (j < len2) {
            merged[k] = arr2[j];
            j++;
            k++;
        }
    }

    // =======================
    // 测试用例函数
    // =======================
    function testExample1() public pure returns (uint256[] memory result) {
        uint256[] memory arr1 = new uint256[](3);
        uint256[] memory arr2 = new uint256[](4);
        arr1[0] = 1; arr1[1] = 3; arr1[2] = 6;
        arr2[0] = 9; arr2[1] = 5; arr2[2] = 9; arr2[2] = 2;

        // [1,2,3,5,6,9]
        result = mergeSorted(arr1, arr2);

        return result;
    }
}
