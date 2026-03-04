package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTClaims 自定义JWT声明
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	RoleID   int    `json:"role_id"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID int, username string, roleID int, secret string, expireHour int) (string, error) {
	// 设置token过期时间
	expireTime := time.Now().Add(time.Hour * time.Duration(expireHour))

	// 创建JWT声明
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "goshopadmin",
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的token字符串
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT token
func ParseToken(tokenString string, secret string) (*JWTClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
