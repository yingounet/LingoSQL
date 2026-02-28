package utils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 校验密码强度（至少8位，包含大小写与数字）
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度至少 8 位")
	}
	hasUpper := false
	hasLower := false
	hasNumber := false
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber {
		return errors.New("密码需包含大小写字母与数字")
	}
	return nil
}
