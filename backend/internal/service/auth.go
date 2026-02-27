package service

import (
	"errors"

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
func (s *AuthService) Register(req *models.UserCreateRequest) (*models.User, string, error) {
	// 检查用户名是否已存在
	exists, err := s.userDAO.ExistsByUsername(req.Username)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	exists, err = s.userDAO.ExistsByEmail(req.Email)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", errors.New("邮箱已存在")
	}

	// 加密密码
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", err
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, "", err
	}

	// 生成 JWT token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login 用户登录
func (s *AuthService) Login(req *models.UserLoginRequest) (*models.User, string, error) {
	// 获取用户
	user, err := s.userDAO.GetByUsername(req.Username)
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 生成 JWT token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GetUser 获取用户信息
func (s *AuthService) GetUser(userID int) (*models.User, error) {
	return s.userDAO.GetByID(userID)
}
