// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import {IERC721Receiver} from "@openzeppelin/contracts/token/ERC721/IERC721Receiver.sol";

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

import {PriceOracle} from "./lib/PriceOracle.sol";
import "hardhat/console.sol";

/**
 * @title NFT拍卖合约
 * @dev 拍卖合约，支持ETH和ERC20代币竞拍，集成 Chainlink 价格预言机
 * 主链：Sepolia测试网
 */
contract NFTAuction is IERC721Receiver, ERC721Holder, Initializable, OwnableUpgradeable, ReentrancyGuardUpgradeable {

    // NFT拍卖信息
    struct AuctionData{
        address seller; // NFT卖家所有者地址
        address nftContract;
        uint256 tokenId;
        uint256 startPrice; // 以 USD 计价（18位小数）
        uint256 startTime;
        uint256 duration;
        uint256 feeRate; // 手续费率（基点 basis points，1% = 100）
    }

    // NFT出价信息
    struct NftBidInfo {
        address highestBidder; // 拍卖的最高出价者
        uint256 usdPrice; // 换算为USD价格
        address paymentToken; // 拍卖的最高出价者付款方式ERC20或ETH address(0) 表示 ETH
        uint256  tokenAmount; // 拍卖的最高出价者付出的ERC20或ETH数量，当为-1时表示没有出价信息，为0是表示跨链竞拍
    }

    bool auctionEnded; // 拍卖是否结束
    bool auctionClaimed;

    uint256 feeRate; // 手续费率（基点 basis points，1% = 100）

    // 状态量
    AuctionData public auctionData;
    NftBidInfo public nftBidInfo;

    // 合约工厂
    address public factory;

    // Chainlink 价格预言机
    PriceOracle public priceOracle;


    // 竞拍出价历史
    // mapping(address => NftBidInfo) public bidHistory;

    // 事件
    event BidPlaced(address indexed bidder, uint256 amount, uint256 usdValue);
    event BidNftReturn(address indexed seller, address indexed nftContract, uint256 tokenId);
    event AuctionEnded(address indexed winner, uint256 amount);
    event AuctionClaimed(address indexed winner, uint256 tokenId);
    event AuctionCancelled(address indexed seller);
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
        uint256 _feeRate,
        address _factory,
        address _priceOracle
    ) external initializer {
        
        console.log("--- NFTAuction.initialize() logs ---");
        console.log("block.timestamp:", block.timestamp);
        console.log("_seller:", _seller);
        console.log("_nftContract:", _nftContract);
        console.log("_tokenId:", _tokenId);
        console.log("_startPriceUSD:", _startPriceUSD);
        console.log("_startTime:", _startTime);
        console.log("_duration:", _duration);
        console.log("_feeRate:", _feeRate);
        console.log("_factory:", _factory);
        console.log("_priceOracle:", _priceOracle);
        console.log("msg.sender (factory):", msg.sender);

        require(_startPriceUSD > 0, "Start price must be > 0");
        // require(_startTime >= block.timestamp, "Start time must be in the future");
        require(_seller != address(0), "Invalid NFT seller address");
        require(_nftContract != address(0), "Invalid NFT contract address");
        //require( IERC721(_nftContract).ownerOf(_tokenId) == address(this), "Auction need received the NFT"); // 拍卖时NFT资产已经转入拍卖合约中
        require(_duration > 0, "Duration must be greater than 0");
        require(_feeRate <= 10000, "Fee rate too high"); // 最高 100%
        // require(_ethUsdPriceFeed != address(0), "Invalid price datafeed address");

        __Ownable_init(msg.sender);
        __ReentrancyGuard_init();

        auctionData = AuctionData({
            seller: _seller,
            nftContract: _nftContract,
            tokenId: _tokenId,
            startPrice: _startPriceUSD,
            startTime: _startTime,
            duration: _duration,
            feeRate: _feeRate
        });

        nftBidInfo = NftBidInfo({
            highestBidder: address(0),
            usdPrice: 0,
            paymentToken:address(0),
            tokenAmount: 0
        });

        auctionEnded=false;
        auctionClaimed=false;

        priceOracle = PriceOracle(_priceOracle);

        factory = _factory;
        console.log("--- Initialization complete ---");
        // 转移NFT到合约
        // IERC721(_nftContract).safeTransferFrom(
        //     _factory,
        //     address(this),
        //     _tokenId
        // );
    }

    /**
     * @dev ETH出价
     */
    function bidEth() external payable nonReentrant {
        console.log("--- NFTAuction.bidEth() logs ---");
        console.log("block.timestamp:", block.timestamp);
        console.log("auctionData startTime:", auctionData.startTime);
        console.log("nftBidInfo usdPrice:", nftBidInfo.usdPrice);
        console.log("auctionData startPrice:", auctionData.startPrice);
        console.log("auctionData duration:", auctionData.duration);

        require(block.timestamp >= auctionData.startTime && block.timestamp <= auctionData.startTime+auctionData.duration, "Auction Not active");
        require(!auctionEnded, "Auction finalized");
        require(msg.value > 0, "ETH should greate than 0");

        uint256 usdValue = priceOracle.calculateUsdValue(msg.value, nftBidInfo.paymentToken) / 1e12; // 换算到拍卖价价格精度，美元单位，下同
        console.log("usdValue:", usdValue);

        require(usdValue >= auctionData.startPrice, "Bid below start price"); // 首先不少于起拍价
        require(usdValue > nftBidInfo.usdPrice, "Bid not high enough"); // 其次大于上一个拍卖出价

        console.log("nftBidInfo current Bidder:", msg.sender);

        // 先退还上一个出价者
        if (nftBidInfo.highestBidder != address(0)) {
            _refundPreviousBidder();
        }

        console.log("NFTAuction Current Balance: ",address(this).balance);

        // 记录本次出价信息
        nftBidInfo.highestBidder = msg.sender;
        nftBidInfo.usdPrice = usdValue;
        nftBidInfo.paymentToken = address(0);
        nftBidInfo.tokenAmount = msg.value;

        emit BidPlaced(msg.sender, msg.value, usdValue);

        console.log("--- NFTAuction.bidEth() complete ---");
    }

    /**
     * @dev ERC20 出价
     * @param amount 代币数量
     * @param paymentToken 代币CA
     */
    function bidERC20(uint256 amount, address paymentToken) external nonReentrant {
        console.log("bidERC20 - amount: ", amount);
        console.log("bidERC20 - paymentToken: ", paymentToken);

        uint256 allowance = IERC20(paymentToken).allowance(msg.sender, address(this));
        require(allowance >= amount, "ERC20: insufficient allowance");

        // require(block.timestamp >= auctionData.startTime && block.timestamp <= auctionData.startTime+auctionData.duration, "Auction Not active");
        require(paymentToken != address(0), "Use bidEth for ETH");
        require(!auctionEnded, "Auction finalized");
        require(amount > 0, "Invalid token amount");

        // 接收ERC20代币
        uint256 balance = IERC20(paymentToken).balanceOf(msg.sender);
        console.log("bidERC20 - user ERC20 Token balance: ", balance);
        require(balance>=amount,
            string(
                abi.encodePacked(
                    "Your ERC20 Token balance should not less than ",
                    uint2str(amount)
                )
            ));
        require(
            IERC20(paymentToken).transferFrom(msg.sender, address(this), amount),
            "Token transfer failed"
        );

        // 计算出价金额
        uint256 usdValue = priceOracle.calculateUsdValue(amount, paymentToken) / 1e12;
        console.log("bidERC20 - ERC20 current usdValue: ", usdValue); // 6位精度，参考USD

        require(usdValue >= auctionData.startPrice, "Bid below start price");
        require(usdValue > nftBidInfo.usdPrice, "Bid not high enough");

        // 先记录ERC20的出价信息
        nftBidInfo.highestBidder = msg.sender;
        nftBidInfo.usdPrice = usdValue;
        nftBidInfo.paymentToken = paymentToken;
        nftBidInfo.tokenAmount = amount;

        // 再退还上一个出价者
        if (nftBidInfo.highestBidder != address(0)) {
            _refundPreviousBidder();
        }

        emit BidPlaced(msg.sender, amount, usdValue);
    }

    /**
     * @notice 结束拍卖
     */
    function endAuction() external {
        require(block.timestamp >= auctionData.startTime+auctionData.duration, "Auction is still ongoing");
        require(!auctionEnded, "Auction Already ended");

        auctionEnded = true;

        claim();

        // NFT拍卖结束
        emit AuctionEnded(nftBidInfo.highestBidder, nftBidInfo.usdPrice);
    }

    /**
    * @notice 卖家取消拍卖（仅在无出价时）
     */
    function cancelAuction() external nonReentrant onlyOwner {
        require(msg.sender == auctionData.seller, "Only NFT Seller Cancel Auction");
        require(auctionEnded, "Auction Already ended");
        require(nftBidInfo.highestBidder != address(0), "Cannot cancel with active bids");

        auctionEnded = true;

        // 退还 NFT
        _nftSafeTransferFrom(auctionData.seller);

        emit AuctionCancelled(auctionData.seller);
    }

    /**
    * @notice 检查拍卖是否可以结束
     */
    function canEnd() external view returns (bool) {
        console.log("canEnd - timestamp",block.timestamp);
        console.log("canEnd - Start Time",auctionData.startTime);
        console.log("canEnd - End TIme",auctionData.startTime + auctionData.duration);
        console.log("canEnd - auctionEnded",auctionEnded);
        return !auctionEnded &&
            block.timestamp >= auctionData.startTime + auctionData.duration;
    }

    /**
   * @notice 获取剩余时间
     */
    function timeRemaining() external view returns (uint256) {
        if (auctionEnded) return 0;
        uint256 endTime = auctionData.startTime + auctionData.duration;
        if (block.timestamp >= endTime) return 0;
        return endTime - block.timestamp;
    }

    /**
    * @notice 领取NFT和资金
     */
    function claim() public {
        require(auctionEnded, "Auction not ended");
        require(!auctionClaimed, "Already claimed");

        auctionClaimed = true;

        if (nftBidInfo.highestBidder != address(0)) {
            // 计算平台手续费
            uint256 fee = (nftBidInfo.tokenAmount * auctionData.feeRate) / 10000;
            uint256 sellerAmount = nftBidInfo.tokenAmount - fee;

            // NFT转移给竞拍出价高者
            _nftSafeTransferFrom(nftBidInfo.highestBidder);

            // 转移资金给NFT卖家
            if (nftBidInfo.paymentToken == address(0)) {
                payable(auctionData.seller).transfer(sellerAmount);
            } else {
                IERC20(nftBidInfo.paymentToken).transfer(auctionData.seller, sellerAmount);
            }

            emit AuctionClaimed(nftBidInfo.highestBidder, auctionData.tokenId);
        } else {
            // 没有竞拍者出价则退还NFT和资金
            _nftSafeTransferFrom(auctionData.seller);
            if (nftBidInfo.paymentToken == address(0)) {
                payable(auctionData.seller).transfer(nftBidInfo.tokenAmount);
            } else {
                IERC20(nftBidInfo.paymentToken).transfer(auctionData.seller, nftBidInfo.tokenAmount);
            }

            // NFT资产退回
            emit BidNftReturn(auctionData.seller, auctionData.nftContract, auctionData.tokenId);
        }
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
     * @dev 退还上一个出价者的资金
     */
    function _refundPreviousBidder() internal{
        console.log("_refundPreviousBidder - last bidder: ", nftBidInfo.highestBidder);
        console.log("_refundPreviousBidder - last amount: ", nftBidInfo.tokenAmount);

        require(nftBidInfo.highestBidder != address(0), "Invalid Bidder address");
        require(nftBidInfo.tokenAmount > 0, "PaymentToken amount Value must be > 0");

        uint256 amount = uint256(nftBidInfo.tokenAmount);

        if (nftBidInfo.paymentToken != address(0)) {
            require(
                IERC20(nftBidInfo.paymentToken).transfer(nftBidInfo.highestBidder, amount),
                "ERC20 Transfer failed"
            );
        } else {
            (bool success, ) = payable(nftBidInfo.highestBidder).call{value: amount}("");
            // console.log("_refundPreviousBidder is success: ", success);
            require(success, "ETH Transfer failed");
        }
    }

    /**
     * @dev NFT资产退还
     */
    function _nftSafeTransferFrom(address recipient) internal onlyOwner{
        require(recipient != address(0), "Invalid NFT Recipient");

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

    function uint2str(uint256 _i) internal pure returns (string memory str) {
        if (_i == 0) {
            return "0";
        }
        uint256 j = _i;
        uint256 len;
        while (j != 0) {
            len++;
            j /= 10;
        }
        bytes memory bstr = new bytes(len);
        uint256 k = len;
        j = _i;
        while (j != 0) {
            bstr[--k] = bytes1(uint8(48 + j % 10));
            j /= 10;
        }
        str = string(bstr);
    }

    // 为将来升级保留存储空间
    uint256[50] private __gap;
}