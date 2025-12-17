// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC721Receiver} from "@openzeppelin/contracts/token/ERC721/IERC721Receiver.sol";

import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

import "./NFTAuction.sol";
import {PriceOracle} from "./lib/PriceOracle.sol";

/**
 * @title AuctionFactory
 * @notice 拍卖工厂合约，采用 Uniswap V2 风格工厂模式
 * @dev 每场拍卖部署独立的 Auction 合约实例，支持 UUPS 升级（Factory 本身也是可升级的）
 */
contract AuctionFactory is IERC721Receiver, Initializable, UUPSUpgradeable, AccessControlUpgradeable, ReentrancyGuardUpgradeable {

    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    address public auctionImplementation;

    address[] public allAuctions;
    mapping(address => address[]) public sellerAuctions;  // 卖家 => 拍卖列表
    mapping(address => mapping(uint256 => address[])) public nftTokenAuctions; // NFT合约 => TokenId => 拍卖列表

    // 手续费配置（基点，10000 = 100%）
    uint256 public feeRate;
    address public feeRecipient;

    PriceOracle public priceOracle;

    // 事件
    event AuctionCreated(
        address indexed auction,
        address indexed seller,
        address nftContract,
        uint256 tokenId,
        uint256 startPriceUSD
    );
    event ImplementationUpgraded(address indexed newImplementation);
    event FeeWithdrawn(address indexed token, address indexed to, uint256 amount);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        // 在remix本地测试非代理模式时，可以注释掉下面这行
        _disableInitializers();
    }

    function initialize(address admin, address _auctionImplementation, address _priceOracle) public initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        _grantRole(OPERATOR_ROLE, admin);

        feeRate = 250; // 2.5%
        feeRecipient = msg.sender;

        auctionImplementation = _auctionImplementation;
        priceOracle = PriceOracle(_priceOracle);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    /**
     * @notice 创建新拍卖
     */
    function createAuction(
        address nftContract,
        uint256 tokenId,
        uint256 startPriceUSD,
        uint256 startTime,
        uint256 duration
    ) external returns (address) {
        require(nftContract != address(0), "Invalid NFT contract address");

        // 获取当前 NFT 所有者地址，只有 NFT 所有者与Seller一致才可以创建拍卖
        address nftOwner = IERC721(nftContract).ownerOf(tokenId);
        require(nftOwner == msg.sender, "You are not the owner of this NFT");

        bool flag1 = IERC721(nftContract).isApprovedForAll(msg.sender, address(this));
        bool flag2 = IERC721(nftContract).getApproved(tokenId) == address(this);

        // 确保 NFT 已经被授权给工厂合约
        require(
            flag1 || flag2,
            "NFT not approved for transfer"
        );

        // 将 NFT 从 seller 转移到 Factory（使用 safeTransferFrom 更安全）
        IERC721(nftContract).safeTransferFrom(nftOwner, address(this), tokenId);

        // 部署代理合约并初始化
        address proxy = address(new ERC1967Proxy(
            auctionImplementation,
            abi.encodeWithSelector(
                NFTAuction.initialize.selector,
                nftOwner,
                nftContract,
                tokenId,
                startPriceUSD,
                startTime,
                duration,
                feeRate,
                address(this),
                address(priceOracle)
            )
        ));

        // 将 NFT 从 Factory 转移到拍卖合约 Auction 中锁定（注意：从 address(this) 转出）
        IERC721(nftContract).safeTransferFrom(address(this), proxy, tokenId);

        // 记录拍卖信息
        allAuctions.push(proxy);
        sellerAuctions[nftOwner].push(proxy);
        nftTokenAuctions[nftContract][tokenId].push(proxy);

        emit AuctionCreated(proxy, nftOwner, nftContract, tokenId, startPriceUSD);

        return proxy;
    }

    /**
     * @notice 升级拍卖实现地址（只是更新 implementation 指针，后续新拍卖会用新实现）
     */
    function upgradeAuctionImplementation(address newImplementation) external onlyRole(UPGRADER_ROLE) {
        auctionImplementation = newImplementation;
        emit ImplementationUpgraded(newImplementation);
    }

    /**
     * @notice 为特定拍卖设置代币价格预言机
     */
    function setAuctionTokenPriceFeed(address token, address otherPriceFeed) external onlyRole(DEFAULT_ADMIN_ROLE) {
        priceOracle.setTokenPriceFeed(token, otherPriceFeed);
    }

    /**
     * @notice 设置手续费率（基点，10000 = 100%）
     * @dev 限制最大为 10%（即 1000）
     */
    function setFeeRate(uint256 _feeRate) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(_feeRate <= 1000, "Fee too high"); // 最大 10%
        feeRate = _feeRate;
    }

    /**
     * @notice 根据场次顺序获取拍卖信息
     */
    function getAuction(uint256 auctionIndexId) external view returns (address) {
        require(auctionIndexId < allAuctions.length, "Invalid auction Index ID");
        return allAuctions[auctionIndexId];
    }

    function getAllAuctions() external view returns (address[] memory) {
        return allAuctions;
    }

    function getSellerAuctions(address seller) external view returns (address[] memory) {
        return sellerAuctions[seller];
    }

    function getAuctionsByNFT(address nftContract, uint256 tokenId) external view returns (address[] memory) {
        return nftTokenAuctions[nftContract][tokenId];
    }

    /**
     * @notice 结束特定场次的拍卖（工厂作为入口，调用具体 Auction 的 endAuction）
     */
    function endAuction(uint256 auctionId) external {
        require(auctionId < allAuctions.length, "Invalid auction ID");
        NFTAuction(allAuctions[auctionId]).endAuction();
    }

    function getAuctionCount() external view returns (uint256) {
        return allAuctions.length;
    }

    /**
     * @notice 提取平台手续费（ETH）
     */
    function withdrawFees(address payable to) external onlyRole(DEFAULT_ADMIN_ROLE) nonReentrant {
        require(to != address(0), "Invalid Address");
        uint256 balance = address(this).balance;
        require(balance > 0, "balance is zero");

        (bool success, ) = to.call{value: balance}("");
        require(success, "Balance Withdraw Failed");

        emit FeeWithdrawn(address(0), to, balance);
    }

    /**
     * @notice 提取 ERC20 手续（如果平台收 ERC20 手续时使用）
     */
    function withdrawERC20Fees(IERC20 token, address to, uint256 amount) external onlyRole(DEFAULT_ADMIN_ROLE) nonReentrant {
        require(to != address(0), "Invalid Address");
        require(amount > 0, "amount zero");
        uint256 bal = token.balanceOf(address(this));
        require(bal >= amount, "insufficient token balance");

        bool ok = token.transfer(to, amount);
        require(ok, "ERC20 transfer failed");

        emit FeeWithdrawn(address(token), to, amount);
    }

    // ERC721 接受函数
    function onERC721Received(
        address,
        address,
        uint256,
        bytes calldata
    ) external pure returns (bytes4) {
        return IERC721Receiver.onERC721Received.selector;
    }
}
