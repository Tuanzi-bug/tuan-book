package service

import (
	"context"
	"errors"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany       = repository.ErrCodeSendTooMany
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
	ErrUserDuplicateEmail    = repository.ErrUserDuplicate
)

//go:generate mockgen -source=./user.go -package=svcmocks -destination=./mocks/user.mock.go UserService
type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx *gin.Context, user domain.User) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 对密码进行加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	// 获取数据库
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	// 检查密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) FindById(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, id)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 无论是有问题还是没问题，都会进入这个。用户找不到情况就会执行下面的创建部分
		return u, err
	}
	ud := domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, ud)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, err
	}
	// 发生冲突就再查一遍
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) UpdateNonSensitiveInfo(ctx *gin.Context, user domain.User) error {
	return svc.repo.UpdateNonSensitiveInfo(ctx, user)
}
