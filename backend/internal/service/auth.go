package service

import (
	"errors"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
)

type AuthService struct {
	userDAO *sqlite.UserDAO
}

func NewAuthService(userDAO *sqlite.UserDAO) *AuthService {
	return &AuthService{userDAO: userDAO}
}

// Register 用户注册
func (s *AuthService) Register(req *models.UserCreateRequest) (*models.User, string, string, error) {
	// 检查用户名是否已存在
	exists, err := s.userDAO.ExistsByUsername(req.Username)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	exists, err = s.userDAO.ExistsByEmail(req.Email)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", errors.New("邮箱已存在")
	}

	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return nil, "", "", err
	}

	// 加密密码
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", "", err
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, "", "", err
	}

	// 生成访问 token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// Login 用户登录
func (s *AuthService) Login(req *models.UserLoginRequest) (*models.User, string, string, error) {
	// 获取用户
	user, err := s.userDAO.GetByUsername(req.Username)
	if err != nil {
		return nil, "", "", errors.New("用户名或密码错误")
	}

	now := time.Now()
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		return nil, "", "", errors.New("账户已锁定，请稍后再试")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		failedCount := 1
		if user.LastFailedLoginAt != nil && now.Sub(*user.LastFailedLoginAt) <= 15*time.Minute {
			failedCount = user.FailedLoginCount + 1
		}
		var lockedUntil *time.Time
		if failedCount >= 5 {
			lockTime := now.Add(15 * time.Minute)
			lockedUntil = &lockTime
		}
		_ = s.userDAO.UpdateLoginFailure(user.ID, failedCount, now, lockedUntil)
		return nil, "", "", errors.New("用户名或密码错误")
	}

	if err := s.userDAO.UpdateLoginSuccess(user.ID); err != nil {
		return nil, "", "", err
	}

	// 生成访问 token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// GetUser 获取用户信息
func (s *AuthService) GetUser(userID int) (*models.User, error) {
	return s.userDAO.GetByID(userID)
}

// RefreshToken 刷新访问令牌
func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	claims, err := utils.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}
	return utils.GenerateAccessToken(claims.UserID, claims.Username)
}
