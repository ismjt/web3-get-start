package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/snluu/uuid"
)

// Md5 计算字符串的md5值
func Md5(source string) string {
	md5h := md5.New()
	md5h.Write([]byte(source))
	return hex.EncodeToString(md5h.Sum(nil))
}

func Truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

func Len(s string) int {
	return len([]rune(s))
}

func UUID() string {
	return uuid.Rand().Hex()
}

func GetCurrentTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc)
}

func GetCurrentDirectory() string {
	//dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//if err != nil {
	//	seelog.Critical(err)
	//}
	//return strings.Replace(dir, "\\", "/", -1)
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	return absPath
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Decrypt(ciphertext, key []byte) ([]byte, error) {
	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		err = errors.New("Text is too short")
		return nil, err
	}

	// Get the 16 byte IV
	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

func Encrypt(plaintext, key []byte) ([]byte, error) {

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// 校验密码强度
func ValidatePasswordStrength(password string) error {
	if len(password) < 6 {
		return errors.New("密码长度至少为6位")
	}

	// 至少一个大写字母
	if ok, _ := regexp.MatchString(`[A-Z]`, password); !ok {
		return errors.New("密码必须包含至少一个大写字母")
	}

	// 至少一个小写字母
	if ok, _ := regexp.MatchString(`[a-z]`, password); !ok {
		return errors.New("密码必须包含至少一个小写字母")
	}

	// 至少一个数字
	if ok, _ := regexp.MatchString(`[0-9]`, password); !ok {
		return errors.New("密码必须包含至少一个数字")
	}

	// 至少一个特殊字符
	if ok, _ := regexp.MatchString(`[!@#\$%\^&\*\(\)\-_=\+\[\]\{\}\|;:'",.<>?/]+`, password); !ok {
		return errors.New("密码必须包含至少一个特殊字符")
	}

	return nil
}

// 从字符集中随机取一个字符
func randomChar(chars string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		return 0, err
	}
	return chars[n.Int64()], nil
}

func RandomEmail() (string, error) {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 随机长度 [6, 12]
	lenN, err := rand.Int(rand.Reader, big.NewInt(7))
	if err != nil {
		return "", err
	}
	length := int(lenN.Int64()) + 6

	// 构造前缀
	prefix := make([]byte, length)
	for i := 0; i < length; i++ {
		c, err := randomChar(chars)
		if err != nil {
			return "", err
		}
		prefix[i] = c
	}

	// 随机选择域名
	domains := []string{"@test.com", "@demo.com", "@example.com"}
	domainIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(domains))))
	if err != nil {
		return "", err
	}

	return string(prefix) + domains[domainIdx.Int64()], nil
}
