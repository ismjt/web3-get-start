// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";

/// @notice Minimal mock of Chainlink AggregatorV3Interface for tests
contract MockV3Aggregator is AggregatorV3Interface {
    int256 public _answer; // 价格
    uint256 private _updatedAt;
    uint80 private _roundId;

    constructor() {
        // 默认 3000 USD，8 decimals， Chainlink 风格
        _answer = 3000 * 1e8;
        _updatedAt = block.timestamp;
        _roundId = 1;
    }

    function setLatestAnswer(int256 answer) external {
        _answer = answer;
        _updatedAt = block.timestamp;
        _roundId++;
    }

    function setUpdatedAt(uint256 updatedAt) external {
        _updatedAt = updatedAt;
    }

    function decimals() external pure override returns (uint8) {
        return 8;
    }

    function description() external pure override returns (string memory) {
        return "Mock Aggregator";
    }

    function version() external pure override returns (uint256) {
        return 1;
    }
    function getRoundData(
        uint80 /* _roundId */
    )
    external
    view
    override
    returns (
        uint80 roundId,
        int256 answer,
        uint256 startedAt,
        uint256 updatedAt,
        uint80 answeredInRound
    )
    {
        return (roundId, _answer, _updatedAt, _updatedAt, _roundId);
    }

    function latestRoundData()
    external
    view
    override
    returns (
        uint80 roundId,
        int256 answer,
        uint256 startedAt,
        uint256 updatedAt,
        uint80 answeredInRound
    )
    {
        return (roundId, _answer, _updatedAt, block.timestamp, _roundId);
    }
}
