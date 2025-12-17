// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {IERC721Receiver} from "@openzeppelin/contracts/token/ERC721/IERC721Receiver.sol";

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";

import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

import {PriceOracle} from "../lib/PriceOracle.sol";

/**
 * @title Auction - 跨链NFT拍卖合约
 * @dev 拍卖合约，支持ETH和ERC20代币竞拍，以及跨链竞拍功能
 * 主链：Sepolia测试网
 * 跨链支持：Polygon Amoy测试网
 */
contract CCIPNFTAuction is IERC721Receiver, Initializable, UUPSUpgradeable, OwnableUpgradeable, ReentrancyGuardUpgradeable {

    // 跨链竞拍记录
    struct CrossChainBid {
        address bidder;        // 跨链竞拍者地址
        uint256 amount;        // 竞拍金额(USD)
        uint64 sourceChain;    // 源链选择器
        bool isWinner;         // 是否为获胜者
    }

    // NFT拍卖信息
    struct AuctionData{
        address seller; // NFT卖家所有者地址
        address nftContract;
        uint256 tokenId;
        uint256 startPrice; // 以 USD 计价（18位小数）
        uint256 startTime;
        uint256 duration;
    }

    // NFT出价信息
    struct NftBidInfo {
        address highestBidder; // 拍卖的最高出价者
        uint256 usdPrice; // 换算为USD价格
        address paymentToken; // 拍卖的最高出价者付款方式ERC20或ETH address(0) 表示 ETH
        int256  tokenAmount; // 拍卖的最高出价者付出的ERC20或ETH数量，当为-1时表示没有出价信息，为0是表示跨链竞拍
    }

    bool auctionEnded; // 拍卖是否结束
    bool auctionClaimed;

    AuctionData public auctionData;
    NftBidInfo public nftBidInfo;

    address public factory;

    // ERC20合约地址
    address public immutable ERC20_TOKEN;

    // CCIP适配器地址
    address public ccipAdapter;
    // 跨链竞拍映射
    mapping(bytes32 => CrossChainBid) public crossChainBids;
    bytes32[] public crossChainBidIds;
    // 跨链竞拍状态
    bool public isWinnerCrossChain;
    bytes32 public winningCrossChainBidId;

    // Chainlink 价格预言机
    PriceOracle public priceOracle;


    // 竞拍出价历史
    // mapping(address => NftBidInfo) public bidHistory;

    event BidPlaced(address indexed bidder, uint256 amount, uint256 usdValue);
    event BidNftReturn(address indexed seller, address indexed nftContract, uint256 tokenId);
    event AuctionEnded(address indexed winner, uint256 amount);
    event AuctionClaimed(address indexed winner, uint256 tokenId);
    event CrossChainBidReceived(
        bytes32 indexed messageId,
        address indexed bidder,
        uint256 amount,
        uint64 sourceChain
    );
    event CrossChainAuctionEnded(
        bytes32 indexed messageId,
        address indexed winner,
        uint256 amount,
        uint64 destinationChain
    );
    event NFTReceived(address operator, address from, uint256 tokenId, bytes data);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address _seller,
        address _nftContract,
        uint256 _tokenId,
        uint256 _startPriceUSD,
        uint256 _startTime,
        uint256 _duration,
        address _factory,
        address _priceOracle
    ) external initializer {
        require(_startTime >= block.timestamp, "Start time must be in the future");
        require(_seller != address(0), "Invalid NFT seller address");
        require(_nftContract != address(0), "Invalid NFT contract address");
        require(IERC721(_nftContract).ownerOf(_tokenId) == _seller, "Invalid NFT or owner");
        require(_duration > 0, "Duration must be greater than 0");
        // require(_ethUsdPriceFeed != address(0), "Invalid price datafeed address");

        __Ownable_init(msg.sender);
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();

        auctionData = AuctionData({
            seller: _seller,
            nftContract: _nftContract,
            tokenId: _tokenId,
            startPrice: _startPriceUSD,
            startTime: _startTime,
            duration: _duration
        });

        nftBidInfo = NftBidInfo({
            highestBidder: address(0),
            usdPrice: 0,
            paymentToken:address(0),
            tokenAmount: -1
        });

        auctionEnded=false;
        auctionClaimed=false;

        priceOracle = PriceOracle(_priceOracle);

        factory = _factory;
    }

    /**
     * @dev 设置CCIP适配器地址 (只有NFT所有者可以设置)
     * @param _ccipAdapter CCIP适配器合约地址
     */
    function setCcipAdapter(address _ccipAdapter) external {
        require(msg.sender == auctionData.seller, "Only NFT owner can set CCIP adapter");
        require(_ccipAdapter != address(0), "Invalid CCIP adapter address");
        ccipAdapter = _ccipAdapter;
    }

    /**
     * @dev 接收跨链竞拍 (只有CCIP适配器可以调用)
     * @param messageId CCIP消息ID
     * @param bidder 跨链竞拍者地址
     * @param usdAmount 竞拍金额(USD)
     * @param sourceChain 源链选择器
     */
    function receiveCrossChainBid(
        bytes32 messageId,
        address bidder,
        uint256 usdAmount,
        uint64 sourceChain
    ) external {
        require(msg.sender == ccipAdapter, "Only CCIP adapter can call this function");
        require(bidder != address(0), "Invalid bidder address");
        require(usdAmount > 0, "Bid amount must be greater than 0");

        // 确保拍卖仍在进行中
        require(
            block.timestamp < auctionData.startTime + auctionData.duration,
            "Auction has expired"
        );

        // 跨链竞拍者不能是NFT所有者
        require(bidder != auctionData.seller, "Seller cannot bid");

        // 执行通用出价验证
        _validateBid(usdAmount);

        // 退还上一个出价者的资金 (如果是本地出价者)
        // TODO 实现跨链退款
        if (nftBidInfo.highestBidder != address(0) && !isWinnerCrossChain) {
            _refundPreviousBidder();
        }

        // 记录跨链竞拍
        crossChainBids[messageId] = CrossChainBid({
            bidder: bidder,
            amount: usdAmount,
            sourceChain: sourceChain,
            isWinner: false
        });
        crossChainBidIds.push(messageId);

        // 更新最高出价信息
        nftBidInfo = NftBidInfo({
            highestBidder: bidder,
            usdPrice: usdAmount,
            paymentToken:address(0), // 标记为跨链竞拍
            tokenAmount: 0 // 跨链竞拍没有本地代币数量
        });

        isWinnerCrossChain = true;
        winningCrossChainBidId = messageId;

        emit CrossChainBidReceived(messageId, bidder, usdAmount, sourceChain);
    }

    /**
     * @dev 通用出价验证逻辑
     * @param usdValue 竞拍价格
     */
    function _validateBid(uint256 usdValue) internal view {
        // 确保地址不是零地址
        require(msg.sender != address(0), "Invalid address");
        // 确保拍卖在指定时间
        require(
            block.timestamp < auctionData.startTime + auctionData.duration,
            "Auction has expired"
        );
        // 卖家不能出价
        require(msg.sender != auctionData.seller, "Seller cannot bid");
        // 确保换算为美元之后的价格大于起拍价格，并且大于最高价+增幅
        uint256 minimumBid = (nftBidInfo.highestBidder == address(0)) ? auctionData.startPrice : nftBidInfo.usdPrice;
        require(
            usdValue >= minimumBid,
            "Bid must be higher than starting price and current highest bid"
        );
    }

    /**
     * @dev 退还上一个出价者的资金
     */
    function _refundPreviousBidder() internal {
        if (nftBidInfo.highestBidder != address(0)) {
            require(nftBidInfo.tokenAmount > 0, "PaymentToken amount Value must be > 0");
            uint256 amount = uint256(nftBidInfo.tokenAmount);

            if (nftBidInfo.paymentToken == ERC20_TOKEN) {
                require(
                    IERC20(ERC20_TOKEN).transfer(nftBidInfo.highestBidder, amount),
                    "ERC20 Transfer failed"
                );
            } else {
                (bool success, ) = payable(nftBidInfo.highestBidder).call{
                        value: amount
                    }("");
                require(success, "ETH Transfer failed");
            }
        }
    }



    /**
     * @dev ETH出价
     */
    function bidEth() external payable nonReentrant {
        require(block.timestamp >= auctionData.startTime && block.timestamp <= auctionData.startTime+auctionData.duration, "Auction Not active");
        require(!auctionEnded, "Auction finalized");
        require(msg.value > 0, "ETH should greate than 0");

        uint256 usdValue = priceOracle.calculateUsdValue(msg.value, nftBidInfo.paymentToken);

        require(usdValue >= auctionData.startPrice, "Bid below start price");
        require(usdValue > nftBidInfo.usdPrice, "Bid not high enough");

        // 退还上一个出价者
        if (nftBidInfo.highestBidder != address(0)) {
            _refundPreviousBidder();
        }

        nftBidInfo.highestBidder = msg.sender;
        nftBidInfo.usdPrice = msg.value;

        emit BidPlaced(msg.sender, msg.value, usdValue);
    }

    /**
     * @dev ERC20 出价
     * @param amount 代币数量
     * @param paymentToken 代币CA
     */
    function bidERC20(uint256 amount, address paymentToken) external payable nonReentrant {
        require(block.timestamp >= auctionData.startTime && block.timestamp <= auctionData.startTime+auctionData.duration, "Auction Not active");
        require(!auctionEnded, "Auction finalized");
        require(amount > 0, "Invalid amount");

        IERC20(paymentToken).transferFrom(msg.sender, address(this), amount);

        uint256 usdValue = priceOracle.calculateUsdValue(amount, nftBidInfo.paymentToken);

        require(usdValue >= auctionData.startPrice, "Bid below start price");
        require(usdValue > nftBidInfo.usdPrice, "Bid not high enough");

        // 退还上一个出价者
        if (nftBidInfo.highestBidder != address(0)) {
            _refundPreviousBidder();
        }

        nftBidInfo.highestBidder = msg.sender;
        nftBidInfo.usdPrice = amount;

        emit BidPlaced(msg.sender, amount, usdValue);
    }

    /**
     * @notice 结束拍卖
     */
    function endAuction() external {
        require(block.timestamp >= auctionData.startTime+auctionData.duration, "Auction is still ongoing");
        require(!auctionEnded, "Already ended");

        auctionEnded = true;

        // 如果没有出价者，将NFT返回给所有者
        if (nftBidInfo.highestBidder == address(0)) {
            _nftSafeTransferFrom(auctionData.seller);
            emit BidNftReturn(auctionData.seller, auctionData.nftContract, auctionData.tokenId);
        } else {
            // 如果获胜者是跨链竞拍者
            if (isWinnerCrossChain) {
                // 标记跨链竞拍为获胜者
                crossChainBids[winningCrossChainBidId].isWinner = true;

                // 触发跨链NFT转移事件 (CCIP适配器会监听此事件)
                emit CrossChainAuctionEnded(
                    winningCrossChainBidId,
                    nftBidInfo.highestBidder,
                    nftBidInfo.usdPrice,
                    crossChainBids[winningCrossChainBidId].sourceChain
                );

                // 注意：跨链情况下，NFT暂时保留在合约中
                // 等待CCIP适配器处理跨链转移
                // 实际的NFT转移将通过 transferNFTToCrossChainWinner 函数完成
            } else {
                // 本地获胜者：直接转移NFT
                _nftSafeTransferFrom(nftBidInfo.highestBidder);
            }
        }

        emit AuctionEnded(nftBidInfo.highestBidder, nftBidInfo.usdPrice);
    }

    function _nftSafeTransferFrom(address recipient) internal {
        require(recipient != address(0), "Invalid Bidder");

        // 转移 NFT 给赢家
        try IERC721(auctionData.nftContract).safeTransferFrom(
            address(this),
            recipient,
            auctionData.tokenId
        ) {
        } catch {
            // 如果 safeTransfer 失败，尝试普通转账
            IERC721(auctionData.nftContract).transferFrom(
                address(this),
                recipient,
                auctionData.tokenId
            );
        }
    }

    /**
     * @notice 领取 NFT和资金
     * TODO 扣除手续费
     */
    function claim() external nonReentrant {
        require(auctionEnded, "Auction not ended");
        require(!auctionClaimed, "Already claimed");

        if (nftBidInfo.highestBidder != address(0)) {
            // NFT转移给竞拍出价高者
            _nftSafeTransferFrom(nftBidInfo.highestBidder);
            // 转移资金给卖家
            if (nftBidInfo.paymentToken == address(0)) {
                payable(auctionData.seller).transfer(nftBidInfo.usdPrice);
            } else {
                IERC20(nftBidInfo.paymentToken).transfer(auctionData.seller, nftBidInfo.usdPrice);
            }

            emit AuctionClaimed(nftBidInfo.highestBidder, auctionData.tokenId);
        } else {
            // 没有竞拍者出价则退还NFT和资金
            _nftSafeTransferFrom(auctionData.seller);
            if (nftBidInfo.paymentToken == address(0)) {
                payable(auctionData.seller).transfer(nftBidInfo.usdPrice);
            } else {
                IERC20(nftBidInfo.paymentToken).transfer(auctionData.seller, nftBidInfo.usdPrice);
            }
        }

        auctionClaimed = true;
    }

    /**
     * @notice 获取拍卖信息
     */
    function getAuctionInfo() external view returns (
        AuctionData memory,
        NftBidInfo memory,
        bool,
        bool
    ) {
        return (
            auctionData,
            nftBidInfo,
            auctionEnded,
            auctionClaimed
        );
    }

    /**
     * @dev 将NFT转移给跨链获胜者 (只有CCIP适配器可以调用)
     * @param winner 获胜者地址
     * @param destinationChain 目标链
     */
    function transferNFTToCrossChainWinner(address winner, uint64 destinationChain) external {
        require(msg.sender == ccipAdapter, "Only CCIP adapter can call this function");
        require(isWinnerCrossChain, "Winner is not cross-chain");
        require(winner == nftBidInfo.highestBidder, "Invalid winner address");

        // 记录目标链信息 (用于日志和验证)
        require(destinationChain > 0, "Invalid destination chain");

        // 这里需要与CCIP适配器协调，实现NFT的跨链转移
        // 在实际实现中，可能需要将NFT锁定在合约中，然后在目标链上铸造对应的NFT

        // 暂时将NFT转移给CCIP适配器处理
        IERC721(auctionData.nftContract).transferFrom(address(this), ccipAdapter, auctionData.tokenId);
    }

    /**
     * @dev 获取跨链竞拍信息
     * @param messageId 消息ID
     * @return bidder 竞拍者地址
     * @return amount 竞拍金额
     * @return sourceChain 源链选择器
     * @return isWinner 是否为获胜者
     */
    function getCrossChainBid(bytes32 messageId) external view returns (
        address bidder,
        uint256 amount,
        uint64 sourceChain,
        bool isWinner
    ) {
        CrossChainBid memory ccbid = crossChainBids[messageId];
        return (ccbid.bidder, ccbid.amount, ccbid.sourceChain, ccbid.isWinner);
    }

    /**
     * @dev 获取所有跨链竞拍ID
     * @return 跨链竞拍ID数组
     */
    function getCrossChainBidIds() external view returns (bytes32[] memory) {
        return crossChainBidIds;
    }

    /**
     * @dev 检查是否为跨链获胜者
     * @return 是否为跨链获胜者
     * @return 对应的消息ID
     */
    function getCrossChainWinnerInfo() external view returns (bool, bytes32) {
        return (isWinnerCrossChain, winningCrossChainBidId);
    }

    function onERC721Received(
        address operator,
        address from,
        uint256 tokenId,
        bytes calldata data
    ) external returns (bytes4) {
        emit NFTReceived(operator, from, tokenId, data);
        return this.onERC721Received.selector;
    }

    // UUPS 升级授权，仅 owner 可升级
    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    // 为将来升级保留存储空间
    uint256[50] private __gap;
}