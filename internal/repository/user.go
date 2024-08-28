package repository

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, c *cache.UserCache) *UserRepository {
	return &UserRepository{dao: dao, cache: c}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (repo *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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
	u = domain.User{
		Id:       ud.Id,
		Email:    ud.Email,
		Password: ud.Password,
	}
	// 存入缓存中
	err = repo.cache.Set(ctx, u)
	if err != nil {
		// 需要打监控
		return domain.User{}, err
	}
	return u, err
}
