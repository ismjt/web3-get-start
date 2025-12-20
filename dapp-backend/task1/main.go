package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// Infura API Key
	INFURA_API_KEY = "4ef555db37c84595b51c231cfe198d7f"
	// Sepolia 测试网络 RPC URL
	SEPOLIA_RPC_URL = "https://sepolia.infura.io/v3/" + INFURA_API_KEY
)

// 查询区块信息
func queryBlock(client *ethclient.Client, blockNumber *big.Int) error {
	ctx := context.Background()

	// 获取区块信息
	block, err := client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("获取区块失败: %v", err)
	}

	fmt.Println("\n========== 区块信息 ==========")
	fmt.Printf("区块号: %s\n", block.Number().String())
	fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
	fmt.Printf("父区块哈希: %s\n", block.ParentHash().Hex())
	fmt.Printf("时间戳: %d\n", block.Time())
	fmt.Printf("交易数量: %d\n", len(block.Transactions()))
	fmt.Printf("Gas Used: %d\n", block.GasUsed())
	fmt.Printf("Gas Limit: %d\n", block.GasLimit())
	fmt.Printf("矿工地址: %s\n", block.Coinbase().Hex())
	fmt.Println("==============================\n")

	// 显示前5笔交易哈希
	if len(block.Transactions()) > 0 {
		fmt.Println("区块中的交易 (最多显示5笔):")
		maxTx := len(block.Transactions())
		if maxTx > 5 {
			maxTx = 5
		}
		for i := 0; i < maxTx; i++ {
			fmt.Printf("  %d. %s\n", i+1, block.Transactions()[i].Hash().Hex())
		}
		fmt.Println()
	}

	return nil
}

// 发送交易
func sendTransaction(client *ethclient.Client, privateKeyHex, toAddressHex string, amountETH float64) error {
	ctx := context.Background()

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %v", err)
	}

	// 获取发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("无法转换公钥")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取 nonce
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return fmt.Errorf("获取 nonce 失败: %v", err)
	}

	// 将 ETH 转换为 Wei (1 ETH = 10^18 Wei)
	value := new(big.Int)
	value.SetString(fmt.Sprintf("%.0f", amountETH*1e18), 10)

	// 获取建议的 Gas Price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("获取 gas price 失败: %v", err)
	}

	// 解析接收方地址
	toAddress := common.HexToAddress(toAddressHex)

	// 设置 Gas Limit (简单转账通常需要 21000 gas)
	gasLimit := uint64(21000)

	// 获取链 ID (Sepolia chain ID = 11155111)
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return fmt.Errorf("获取链 ID 失败: %v", err)
	}

	// 构造交易
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return fmt.Errorf("签名交易失败: %v", err)
	}

	// 发送交易
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return fmt.Errorf("发送交易失败: %v", err)
	}

	fmt.Println("\n========== 交易信息 ==========")
	fmt.Printf("发送方地址: %s\n", fromAddress.Hex())
	fmt.Printf("接收方地址: %s\n", toAddress.Hex())
	fmt.Printf("转账金额: %.6f ETH\n", amountETH)
	fmt.Printf("Gas Price: %s Wei\n", gasPrice.String())
	fmt.Printf("Gas Limit: %d\n", gasLimit)
	fmt.Printf("Nonce: %d\n", nonce)
	fmt.Printf("链 ID: %s\n", chainID.String())
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())
	fmt.Println("==============================\n")

	fmt.Printf("在 Sepolia 浏览器查看交易: https://sepolia.etherscan.io/tx/%s\n\n", signedTx.Hash().Hex())

	return nil
}

// 查询账户余额
func queryBalance(client *ethclient.Client, addressHex string) error {
	ctx := context.Background()

	address := common.HexToAddress(addressHex)
	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return fmt.Errorf("查询余额失败: %v", err)
	}

	// 将 Wei 转换为 ETH
	balanceETH := new(big.Float)
	balanceETH.SetString(balance.String())
	ethValue := new(big.Float).Quo(balanceETH, big.NewFloat(1e18))

	fmt.Println("\n========== 账户余额 ==========")
	fmt.Printf("地址: %s\n", address.Hex())
	fmt.Printf("余额: %s ETH\n", ethValue.String())
	fmt.Println("==============================\n")

	return nil
}

func main() {
	// 创建客户端连接
	client, err := ethclient.Dial(SEPOLIA_RPC_URL)
	if err != nil {
		log.Fatalf("连接 Sepolia 网络失败: %v", err)
	}
	defer client.Close()

	fmt.Println("成功连接到 Sepolia 测试网络!")

	// 获取最新区块号
	ctx := context.Background()
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("获取最新区块号失败: %v", err)
	}
	fmt.Printf("当前最新区块号: %d\n", blockNumber)

	// 任务1: 查询最新区块
	fmt.Println("\n--- 任务1: 查询最新区块 ---")
	err = queryBlock(client, big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Printf("查询区块失败: %v\n", err)
	}

	// 任务2: 查询指定区块 (例如区块号 5000000)
	fmt.Println("\n--- 任务2: 查询指定区块 (5000000) ---")
	err = queryBlock(client, big.NewInt(5000000))
	if err != nil {
		log.Printf("查询区块失败: %v\n", err)
	}

	// 任务3: 查询账户余额
	// 注意：替换为你的实际地址
	testAddress := "0xa298831ad84d1a92cf99d744c4c1c827ee06a49a"
	fmt.Println("\n--- 任务3: 查询账户余额 ---")
	err = queryBalance(client, testAddress)
	if err != nil {
		log.Printf("查询余额失败: %v\n", err)
	}

	// 任务4: 发送交易
	// 警告：不在代码中硬编码私钥！生产环境应使用环境变量或密钥管理服务
	// 使用环境变量获取私钥
	privateKey := os.Getenv("SEPOLIA_PRIVATE_KEY")
	toAddress := os.Getenv("SEPOLIA_TO_ADDRESS")

	if privateKey != "" && toAddress != "" {
		fmt.Println("\n--- 任务4: 发送交易 ---")
		// 发送 0.001 ETH
		err = sendTransaction(client, privateKey, toAddress, 0.00001)

		if err != nil {
			log.Printf("发送交易失败: %v\n", err)
		}
	} else {
		fmt.Println("\n--- 任务4: 发送交易 (跳过) ---")
		fmt.Println("提示: 设置环境变量 SEPOLIA_PRIVATE_KEY 和 SEPOLIA_TO_ADDRESS 来发送交易")
		fmt.Println("示例:")
		fmt.Println("  export SEPOLIA_PRIVATE_KEY=your_private_key_without_0x_prefix")
		fmt.Println("  export SEPOLIA_TO_ADDRESS=0x...")
	}

	/**
	 * 本次测试TX：
	 * https://sepolia.etherscan.io/tx/0x2bc37bc7512500ebc704080ff98276f0670128aa9c0b78d49f8b96c863a40908
	 * 测试账户0xc1a65acb1c15c964c34e5aa0f927b23caeed3180余额增加了0.00001SepoliaETH
	 */
	fmt.Println("\n程序执行完成!")
}
