package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Account struct {
	ID      uint `gorm:"primaryKey"`
	Name    string
	Balance float64
}

func (Account) TableName() string {
	return "accounts"
}

type Transaction struct {
	ID            uint    `gorm:"primaryKey"`
	FromAccountId uint    `json:"fromAccountId"` // 转出账户ID
	ToAccountId   uint    `json:"toAccountId"`   // 转入账户ID
	Amount        float64 `json:"amount"`        // 转账金额
}

func (Transaction) TableName() string {
	return "transactions"
}

// Transfer 从 fromAccountId 向 toAccountId 转账 amount 元
func Transfer(db *gorm.DB, fromAccountId, toAccountId uint, amount float64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var fromAcc, toAcc Account
		// 确保转账双方账户存在
		if err := tx.First(&fromAcc, fromAccountId).Error; err != nil {
			return err
		}
		if err := tx.First(&toAcc, toAccountId).Error; err != nil {
			return err
		}

		// 转账条件
		if fromAcc.Balance < amount {
			return fmt.Errorf("账户ID %d 余额不足", fromAccountId)
		}

		// 双方账户余额操作
		if err := tx.Model(&fromAcc).Update("balance", fromAcc.Balance-amount).Error; err != nil {
			return err
		}
		if err := tx.Model(&toAcc).Update("balance", toAcc.Balance+amount).Error; err != nil {
			return err
		}

		// 记录转账事务
		trx := Transaction{FromAccountId: fromAccountId, ToAccountId: toAccountId, Amount: amount}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}
		return nil
	})
}

func RandomFloat64(min, max float64) float64 {
	val := min + rand.Float64()*(max-min)
	return math.Round(val*100) / 100
}

func main() {
	// 1. 连接 SQLite 数据库
	db, err := gorm.Open(sqlite.Open("bank.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 2. 自动迁移表结构
	if err := db.AutoMigrate(&Account{}, &Transaction{}); err != nil {
		log.Fatal(err)
	}

	// 3. 清空数据便于测试
	db.Where("1 = 1").Delete(&Account{})
	db.Where("1 = 1").Delete(&Transaction{})

	// 4. 填充测试数据
	aID := uint(1)
	bID := uint(2)
	amount := 100.00
	randomAmount := RandomFloat64(50, 150) // 生成一个 50~150 之间的随机余额
	var aMockAccount Account
	aResult := db.FirstOrCreate(&aMockAccount, &Account{ID: aID}, &Account{Name: "A", Balance: randomAmount})
	db.FirstOrCreate(&Account{ID: bID}, &Account{Name: "B", Balance: 0})
	if aResult.Error != nil {
		fmt.Printf("账户A模拟数据 - 操作失败:%v\n", aResult.Error)
	} else {
		fmt.Printf("账户A模拟数据 - 操作成功：%+v\n", aMockAccount)
	}

	fmt.Printf("转账前 - 账户%v余额：%v元，账户%v余额：%v元\n", aID, aMockAccount.Balance, bID, 0)

	// 5. 执行转账过程
	if err := Transfer(db, aID, bID, amount); err != nil {
		fmt.Printf("%v转账%v - 失败: %v\n", aID, bID, err)
	} else {
		fmt.Printf("%v转账%v - 成功，金额：%v\n", aID, bID, amount)
	}

	// 6. 查看账户余额
	var accounts []Account
	db.Find(&accounts)
	for _, acc := range accounts {
		fmt.Printf("账户ID:%d, 余额:%.2f\n", acc.ID, acc.Balance)
	}
	// 7. 查询事务记录
	var transactions []Transaction
	db.Find(&transactions)
	for _, tcc := range transactions {
		fmt.Printf("事务记录ID:%d, 转账方:%v, 接收方:%v\n", tcc.ID, tcc.FromAccountId, tcc.ToAccountId)
	}
}
