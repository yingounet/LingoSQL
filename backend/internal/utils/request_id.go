package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRequestID 生成请求 ID
func GenerateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
