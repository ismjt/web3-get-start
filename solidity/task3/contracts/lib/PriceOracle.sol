// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import "@openzeppelin/contracts/access/Ownable.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";

import "hardhat/console.sol";

contract PriceOracle is Ownable {
    AggregatorV3Interface internal ethUsd;
    AggregatorV3Interface internal linkUsd;
    address admin;

    mapping(address => AggregatorV3Interface) public tokenUsdPriceFeeds;

    /**
     * Network:  Arbitrum Sepolia
     * Aggregator: ETH/USD
     * Address: 0xd30e2101a97dcbAeBCBC04F14C3f624E67A35165
     */
    constructor() Ownable(msg.sender) {
        ethUsd = AggregatorV3Interface(
        // ETH/USD sopelia 0x694AA1769357215DE4FAC081bf1f309aDC325306
            0x694AA1769357215DE4FAC081bf1f309aDC325306
        );
        linkUsd = AggregatorV3Interface(
        // LINK/USD sopelia 0xc59E3633BAAC79493d908e63626716e204A45EdF
            0xc59E3633BAAC79493d908e63626716e204A45EdF
        );

        admin = msg.sender;

        tokenUsdPriceFeeds[address(0)] = ethUsd;
        tokenUsdPriceFeeds[0x779877A7B0D9E8603169DdbD7836e478b4624789] = linkUsd;
    }

    // 获取最新的 ETH/USD 价格
    function getLatestPrice() public view returns (int256) {
        (, int256 price, , uint256 updatedAt, ) = ethUsd.latestRoundData();
        require(updatedAt >= block.timestamp - 1 hours, "Data is too old");
        return price;
    }

    // 将 ETH 转换为 USD
    function convertEthToUsd(uint256 ethAmount) public view returns (uint256) {
        int256 ethPrice = getLatestPrice();
        return (ethAmount * uint256(ethPrice)) / 1e8; // 1e8 是 Chainlink 数据的精度
    }

    // 获取最新的 LINK/USD 价格
    function getLatestLinkPrice() public view returns (int256) {
        (, int256 price, , uint256 updatedAt, ) = linkUsd.latestRoundData();
        require(updatedAt >= block.timestamp - 1 hours, "Data is too old");
        return price;
    }

    // 将 LINK 转换为 USD
    function convertLinkToUsd(
        uint256 linkAmount
    ) public view returns (uint256) {
        int256 linkPrice = getLatestLinkPrice();
        return (linkAmount * uint256(linkPrice)) / 1e8; // 1e8 是 Chainlink 数据的精度
    }

    /**
     * @notice 获取 ETH/USD 价格
     */
    function getEthPrice() public view returns (uint256) {
        (, int256 price, , ,) = ethUsd.latestRoundData();
        uint8 decimals = ethUsd.decimals();
        require(price > 0, "Invalid price");
        return _normalizeTo18Decimals(price, decimals);
    }

    /**
     * @notice 获取 ERC20/USD 价格
     */
    function getTokenPrice(address token) public view returns (uint256) {
        AggregatorV3Interface priceFeed = tokenUsdPriceFeeds[token];
        require(address(priceFeed) != address(0), "Price feed not set");
        (, int256 price, , ,) = priceFeed.latestRoundData();
        uint8 decimals = priceFeed.decimals();
        console.log("ERC20 Toekn price: ");
        console.logInt(price);
        require(price > 0, "Invalid price");
        return _normalizeTo18Decimals(price, decimals);
    }

    /**
     * @notice 设置 ERC20 代币价格预言机
     */
    function setTokenPriceFeed(address token, address priceFeed) public onlyOwner{
        // require(msg.sender == admin, "Only Owner can Update");
        require(token != address(0), "ERC20 address invalid");
        require(priceFeed != address(0), "priceFeed address invalid");
        tokenUsdPriceFeeds[token] = AggregatorV3Interface(priceFeed);
    }

    /**
     * @notice 计算出价的 USD 价值
     */
    function calculateUsdValue(uint256 amount, address paymentToken) public view returns (uint256) {
        uint8 decimals = paymentToken == address(0) ? 18 : IERC20Metadata(paymentToken).decimals();
        console.log("calculateUsdValue - token decimals is: ", decimals);
        if (paymentToken == address(0)) {
            // ETH 出价
            uint256 ethPrice = getEthPrice();
            console.log("ETH - calculateUsdValue ethPrice: ", ethPrice);
            console.log("ETH - amount", amount);
            console.log("ETH - paymentToken", paymentToken);
            return (amount * ethPrice) / (10 ** decimals);
        } else {
            // ERC20 出价
            uint256 tokenPrice = getTokenPrice(paymentToken);
            console.log("ERC20 - calculateUsdValue tokenPrice: ", tokenPrice);
            console.log("ERC20 - amount", amount);
            console.log("ERC20 - paymentToken", paymentToken);
            return (amount * tokenPrice) / (10 ** decimals);
        }
    }

    function _normalizeTo18Decimals(int256 price, uint8 feedDecimals) internal pure returns (uint256) {
        if (feedDecimals > 18) {
            return uint256(price) / (10 ** (feedDecimals - 18));
        } else {
            return uint256(price) * (10 ** (18 - feedDecimals));
        }
    }

    function updateAggregator(address aggregator) public onlyOwner {
        require(msg.sender == admin, "Only Owner can Update");
        ethUsd = AggregatorV3Interface(aggregator);
    }
}