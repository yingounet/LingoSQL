package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"lingosql/internal/config"
)

var jwtSecret []byte

// Claims JWT 声明
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// InitJWT 初始化 JWT
func InitJWT() {
	cfg := config.GetConfig()
	jwtSecret = []byte(cfg.JWT.Secret)
}

// GenerateToken 生成 JWT token
func GenerateToken(userID int, username string) (string, error) {
	cfg := config.GetConfig()
	
	expiresIn, err := time.ParseDuration(cfg.JWT.ExpiresIn)
	if err != nil {
		expiresIn = 720 * time.Hour // 默认30天
	}

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析 JWT token
func ParseToken(tokenString string) (*Claims, error) {
	if jwtSecret == nil {
		return nil, errors.New("JWT 未初始化")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的 token")
}
