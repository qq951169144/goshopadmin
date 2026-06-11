package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims 自定义JWT声明
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	RoleID   int    `json:"role_id"`
	jwt.RegisteredClaims
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
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
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
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
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

// CustomerClaims C端客户JWT声明
type CustomerClaims struct {
	CustomerID int `json:"customer_id"`
	jwt.RegisteredClaims
}

// GenerateCustomerToken 生成C端JWT token
func GenerateCustomerToken(customerID int, secret string, expireHour int) (string, error) {
	expireTime := time.Now().Add(time.Hour * time.Duration(expireHour))

	claims := CustomerClaims{
		CustomerID: customerID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "shop-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseCustomerToken 解析C端JWT token
func ParseCustomerToken(tokenString string, secret string) (*CustomerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomerClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomerClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
