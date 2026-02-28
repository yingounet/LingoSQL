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
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// InitJWT 初始化 JWT
func InitJWT() {
	cfg := config.GetConfig()
	jwtSecret = []byte(cfg.JWT.Secret)
}

// GenerateAccessToken 生成访问令牌
func GenerateAccessToken(userID int, username string) (string, error) {
	return generateToken(userID, username, "access", parseDurationOrDefault(config.GetConfig().JWT.ExpiresIn, 24*time.Hour))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID int, username string) (string, error) {
	return generateToken(userID, username, "refresh", parseDurationOrDefault(config.GetConfig().JWT.RefreshExpiresIn, 168*time.Hour))
}

func generateToken(userID int, username, tokenType string, expiresIn time.Duration) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseAccessToken 解析访问令牌
func ParseAccessToken(tokenString string) (*Claims, error) {
	claims, err := parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, errors.New("无效的访问令牌")
	}
	return claims, nil
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string) (*Claims, error) {
	claims, err := parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, errors.New("无效的刷新令牌")
	}
	return claims, nil
}

func parseToken(tokenString string) (*Claims, error) {
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

func parseDurationOrDefault(value string, defaultDuration time.Duration) time.Duration {
	if value == "" {
		return defaultDuration
	}
	expiresIn, err := time.ParseDuration(value)
	if err != nil {
		return defaultDuration
	}
	return expiresIn
}
