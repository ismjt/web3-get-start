package helpers

import (
	"errors"
	"go-blog/models"
	"go-blog/system"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 payload
type MyClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 实现自定义验证接口
func (m MyClaims) Validate() error {
	cfg := system.GetConfiguration()
	issuer, err := m.RegisteredClaims.GetIssuer()
	if err != nil {
		return err
	}
	if issuer != cfg.JWT.Issuer {
		return errors.New("jwt issuer invalid")
	}
	return nil
}

// 生成 JWT token
func GenerateToken(user models.User, exp time.Time) (string, error) {
	cfg := system.GetConfiguration()
	jwtSecret := []byte(cfg.JWT.SK)
	// exp := time.Now().Add(time.Hour * 24)
	claims := MyClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.JWT.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

// 解析并验证 token
func ParseToken(tokenString string) (*MyClaims, error) {
	claims := &MyClaims{}
	cfg := system.GetConfiguration()
	jwtSecret := []byte(cfg.JWT.SK)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 确保签名算法正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err // 解析失败或签名无效
	}

	// Token 无效；token.Valid 表示签名和注册声明都合法；
	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, err
	}

	// 可选：手动验证时间等注册声明
	if err := claims.Validate(); err != nil {
		return nil, err
	}

	return claims, nil
}
