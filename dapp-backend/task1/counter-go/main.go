package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// 引入 abigen 生成的绑定代码
// 注意：counter.go 必须在同一目录或正确包路径下
// 这里假设 counter.go 在当前目录，且 package 为 main

func main() {
	// 加载 .env 文件（仅在开发时需要）
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found") // 生产环境可忽略
	}

	infuraApiKey := os.Getenv("INFURA_API_KEY")
	// 1. 连接到 Sepolia 节点（使用 Infura 或 Alchemy）
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/" + infuraApiKey)
	if err != nil {
		log.Fatal("Failed to connect to Sepolia:", err)
	}
	defer client.Close()

	// 2. 合约地址（替换为你部署的实际地址）
	contractAddress := common.HexToAddress(os.Getenv("CONTRACT_ADDRESS"))
	// contractAddress := common.HexToAddress("0x9E4fb7778AeCC2dB7B3cADd91c4d3A82b177Bb20")

	// 3. 实例化合约
	counter, err := NewCounter(contractAddress, client)
	if err != nil {
		log.Fatal("Failed to instantiate contract:", err)
	}

	ctx := context.Background()

	// 4. 读取当前计数值
	count, err := counter.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatal("Failed to get count:", err)
	}
	fmt.Printf("Current count: %d\n", count)

	// 5. 发送交易：调用 inc()
	privateKeyHex := os.Getenv("SEPOLIA_PRIVATE_KEY") // 不带 0x
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal("Invalid private key:", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatal("Failed to get nonce:", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal("Failed to get gas price:", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111)) // Sepolia chain ID
	if err != nil {
		log.Fatal("Failed to create transactor:", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	tx, err := counter.Inc(auth)
	if err != nil {
		log.Fatal("Failed to send increment transaction:", err)
	}

	// 实际测试结果：0x7dd60278155975ff78af597ebea6cad52867e1380200b1151a50cad9e67224f5
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())

	// 6. 等待交易确认（可选，等待时间比较长）
	_, err = bind.WaitMined(ctx, client, tx)
	if err != nil {
		log.Println("Warning: transaction not mined yet, but may succeed later")
	} else {
		fmt.Println("Transaction confirmed!")
	}

	// 7. 再次读取计数值
	newCount, err := counter.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatal("Failed to get new count:", err)
	}
	fmt.Printf("New count after increment: %d\n", newCount)
}
