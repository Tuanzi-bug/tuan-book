package repository

import (
	"context"
	"database/sql"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/gin-gonic/gin"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	UpdateNonSensitiveInfo(ctx *gin.Context, user domain.User) error
}

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCacheUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository {
	return &CacheUserRepository{dao: dao, cache: c}
}

func (repo *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(u))
}

func (repo *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(u), nil
}

func (repo *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从缓存获取数据
	u, err := repo.cache.Get(ctx, id)
	if err == nil {
		return u, err
	}
	// 从数据库中获取数据
	ud, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = repo.entityToDomain(ud)
	// 存入缓存中
	err = repo.cache.Set(ctx, u)
	if err != nil {
		// 需要打监控
		return domain.User{}, err
	}
	return u, err
}

func (repo *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(u), nil
}

func (repo *CacheUserRepository) UpdateNonSensitiveInfo(ctx *gin.Context, user domain.User) error {
	return repo.dao.UpdateById(ctx, repo.domainToEntity(user))
}

func (repo *CacheUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
		Nickname: u.Nickname,
		Ctime:    u.Ctime.UnixMilli(),
	}
}

func (repo *CacheUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
