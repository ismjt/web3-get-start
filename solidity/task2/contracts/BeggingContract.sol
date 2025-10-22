// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

contract BeggingContract is Ownable {
    using Strings for uint256;

    // 记录捐赠者的地址和金额
    mapping(address => uint256) public donations;

    // 前 3 名捐赠者
    address[3] public topDonors;

    // 时间限制：捐赠开始和结束时间（UNIX 时间戳）
    uint256 public donationStart;
    uint256 public donationEnd;

    // 事件：记录捐赠和提款
    event Log(string msg);
    event Donated(address indexed donor, uint256 amount);
    event Withdrawn(address indexed owner, uint256 amount);

    constructor(uint256 _start, uint256 _end) Ownable(msg.sender){
        require(_start < _end, "Start time must be before end time");
        donationStart = _start; // 示例：1761122000
        donationEnd = _end; // 示例：1761142999
    }

    // 修饰符：检查当前时间是否在允许捐赠区间
    modifier onlyDuringDonationPeriod() {
        require(
            block.timestamp >= donationStart && block.timestamp <= donationEnd,
            "Donations not allowed at this time"
        );
        _;
    }

    /// @notice 捐赠函数，允许用户在指定期限内发送 ETH 给合约
    /// @dev 使用 payable 以便接收 ETH
    function donate() external payable onlyDuringDonationPeriod {
        require(msg.value > 0, "You need send some ETH to donate");

        donations[msg.sender] += msg.value;
        emit Donated(msg.sender, msg.value);

        _updateTopDonors(msg.sender);
    }

    /// @notice 查询某个地址的捐赠总额
    /// @param donor 需要查询的捐赠者地址
    /// @return amount 该地址的累计捐赠金额（单位：wei）
    function getDonation(address donor) external view returns (uint256 amount) {
        return donations[donor];
    }

    /// @notice 提取所有捐赠资金（仅限合约拥有者）
    function withdraw() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No funds to withdraw");

        // 转账给拥有者
        (bool success, ) = owner().call{value: balance}("");
        require(success, "Withdrawal failed");

        emit Withdrawn(owner(), balance);
    }

    /// @notice 返回合约当前余额
    function getBalance() external view returns (uint256) {
        return address(this).balance;
    }

    /// @notice 获取TOP3捐赠排行
    function getTopDonors() external view returns (address[3] memory) {
        return topDonors;
    }

    /// @notice 更新捐赠排行榜
    function _updateTopDonors(address donor) internal {
        for (uint i = 0; i < 3; i++) {
            if (topDonors[i] == donor) break;
        }

        for (uint i = 0; i < 3; i++) {
            if (topDonors[i] == address(0) || donations[donor] > donations[topDonors[i]]) {
                for (uint j = 2; j > i; j--) {
                    topDonors[j] = topDonors[j - 1];
                }
                topDonors[i] = donor;
                break;
            }
        }
    }

    /// @notice 更新捐赠起止时间
    function updateDonationTm(uint256 _start, uint256 _end) public onlyOwner{
        emit Log(string.concat("timestamp: ",block.timestamp.toString()));
        donationStart = _start;
        donationEnd = _end;

        emit Log(string.concat("[UPDATE]Donation Start: ",donationStart.toString(),",  Donation End: ", donationEnd.toString()));
    }


    // fallback & receive 用于接收直接转账（无调用 donate）
    receive() external payable onlyDuringDonationPeriod {
        donations[msg.sender] += msg.value;
        emit Donated(msg.sender, msg.value);
    }

    fallback() external payable onlyDuringDonationPeriod {
        donations[msg.sender] += msg.value;
        emit Donated(msg.sender, msg.value);
    }
}