package main

import (
	"fmt"
	"sort"
)

// 136. 只出现一次的数字：给定一个非空整数数组，除了某个元素只出现一次以外，其余每个元素均出现两次。找出那个只出现了一次的元素。可以使用 for 循环遍历数组，结合 if 条件判断和 map 数据结构来解决，例如通过 map 记录每个元素出现的次数，然后再遍历 map 找到出现次数为1的元素。
func findSingleNumber(nums []int) int {
	m := make(map[int]int)
	// 统计每个元素的出现次数
	for _, v := range nums {
		m[v]++
	}
	// 找到出现次数为 1 的元素
	for k, v := range m {
		if v == 1 {
			return k
		}
	}
	// 如果不存在，返回 -1（按题意不会发生）
	return -1
}

// 回文数
// https://leetcode.cn/problems/palindrome-number/description/
func isPalindrome(x int) bool {
	if x < 0 || (x%10 == 0 && x != 0) {
		return false
	}
	// 已经反转的后半部分
	reverted := 0
	for x > reverted {
		reverted = reverted*10 + x%10
		x /= 10
	}
	// 当数字长度为奇数时，reverted/10 忽略入参的中间数字
	return x == reverted || x == reverted/10
}

// 有效的括号
// https://leetcode.cn/problems/valid-parentheses/description/
func isValidStr(s string) bool {
	stack := []rune{}
	brackets := map[rune]rune{')': '(', ']': '[', '}': '{'} // 对应关系

	for _, c := range s {
		switch c {
		case '(', '[', '{':
			stack = append(stack, c) // 左括号入栈
		case ')', ']', '}':
			if len(stack) == 0 || stack[len(stack)-1] != brackets[c] {
				return false // 栈空或者栈顶要素不匹配
			}
			stack = stack[:len(stack)-1] // 弹出栈顶
		default:
			return false // 出现非法字符
		}
	}

	return len(stack) == 0 // 栈空说明全部匹配
}

// 最长公共前缀
// https://leetcode.cn/problems/longest-common-prefix/description/
func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	// 根据字符串长度排序
	sort.Slice(strs, func(i, j int) bool {
		return len(strs[i]) < len(strs[j])
	})
	if len(strs[0]) == 0 {
		return ""
	}
	// 数组中至少两个时
	maxIndex := -1
	for i, b := range strs[0] {
		flag := true
		for _, s := range strs[1:] {
			if string(s[i]) == string(b) {

			} else {
				flag = false
			}
		}
		if flag {
			maxIndex = i
		}
	}
	if maxIndex > -1 {
		return strs[0][:maxIndex+1]
	}
	return ""
}

// 加一
// https://leetcode.cn/problems/plus-one/
func plusOne(digits []int) []int {
	n := len(digits)

	for i := n - 1; i >= 0; i-- {
		if digits[i] < 9 {
			digits[i]++
			return digits
		}
		digits[i] = 0 // 当前位 9+1=10，置0，进位到前一位
	}

	// 遍历完仍有进位，比如 [9,9,9] → [1,0,0,0]
	result := make([]int, n+1)
	result[0] = 1
	return result
}

// 26. 删除有序数组中的重复项
// https://leetcode.cn/problems/remove-duplicates-from-sorted-array/description/
func removeDuplicates(nums []int) int {
	i := 0
	for j := 1; j < len(nums); j++ {
		if nums[i] != nums[j] {
			nums[i+1] = nums[j]
			i++
		}
	}
	return len(nums[:i+1])
}

// 56. 合并区间
// https://leetcode.cn/problems/merge-intervals/description/
func mergeIntervals(intervals [][]int) [][]int {
	// 先对区间数组按照区间的起始位置进行从小到大排序
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})
	arr := intervals[:1]
	for i := 1; i < len(intervals); i++ {
		if intervals[i][0] > arr[len(arr)-1][1] {
			arr = append(arr, intervals[i])
		} else if intervals[i][0] <= arr[len(arr)-1][1] {
			arr[len(arr)-1][1] = intervals[i][1]
		}
	}
	return arr
}

// 两数之和
// https://leetcode-cn.com/problems/two-sum/
func twoSum(nums []int, target int) []int {
	// 假设每种输入只会对应一个答案，并且你不能使用两次相同的元素
	// key: 数值, value: 下标
	m := make(map[int]int)
	for i, num := range nums {
		d := target - num
		if j, ok := m[d]; ok {
			return []int{j, i} // 找到答案
		}
		m[num] = i // 记录当前数值的下标
	}

	return nil // 根据题意，必有答案，这行理论上不会执行
}

func main() {
	fmt.Println("136. 只出现一次的数字:")
	test1 := []int{4, 1, 2, 1, 2, 6, 8, 7, 5, 6, 5, 7, 4}
	fmt.Println("输入数组:", test1)
	result := findSingleNumber(test1)
	fmt.Println("只出现一次的元素是:", result)

	fmt.Println("\n回文数")
	test2 := 12343210
	fmt.Println("回文数输入数据:", test2, "，判断结果：", isPalindrome(test2))

	fmt.Println("\n有效的括号")
	test3 := []string{
		"()", "()[]{}", "(]", "([)]", "{[]}", "", "((({{{[[[]]]}}})))",
	}
	for _, t := range test3 {
		fmt.Printf("输入字符串%s 判断结果 %v\n", t, isValidStr(t))
	}

	fmt.Println("\n最长公共前缀")
	test4 := []string{"flower", "flow", "flight"}
	fmt.Printf("输入参数%s\n输出结果 %v\n", test4, longestCommonPrefix(test4))

	fmt.Println("\n加一")
	test5 := []int{4, 1, 2, 9, 8}
	fmt.Printf("输入参数%v\n", test5)
	fmt.Printf("输出结果 %v\n", plusOne(test5))

	fmt.Println("\n26. 删除有序数组中的重复项")
	test6 := []int{0, 1, 3, 5, 8, 9, 9, 11}
	fmt.Printf("输入参数%v\n", test6)
	fmt.Printf("输出结果 %v\n", removeDuplicates(test6))

	fmt.Println("\n56. 合并区间")
	test7 := [][]int{
		{1, 3},
		{2, 6},
		{8, 10},
		{15, 18},
	}
	fmt.Printf("输入参数%v\n", test7)
	fmt.Printf("输出结果 %v\n", mergeIntervals(test7))

	fmt.Println("\n两数之和")
	test8 := []int{2, 7, 11, 15}
	test9 := 26
	fmt.Printf("输入参数1: %v 参数2: %v\n", test8, test9)
	fmt.Printf("输出结果 %v\n", twoSum(test8, test9))
}
