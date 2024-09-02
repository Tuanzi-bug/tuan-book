package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	cachemocks "github.com/Tuanzi-bug/tuan-book/internal/repository/cache/mocks"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	daomocks "github.com/Tuanzi-bug/tuan-book/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCacheUserRepository_FindById(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)
	testcases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)

		// 固定输入
		ctx context.Context
		// 预期输入
		uid int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中,从数据库中查找",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				urepo := daomocks.NewMockUserDAO(ctrl)
				crepo := cachemocks.NewMockUserCache(ctrl)
				crepo.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				urepo.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{
					Id: uid,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: 100,
					AboutMe:  "自我介绍",
					Phone: sql.NullString{
						String: "15212345678",
						Valid:  true,
					},
					Ctime: nowMs,
					Utime: 102,
				}, nil)
				crepo.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: time.UnixMilli(100),
					AboutMe:  "自我介绍",
					Phone:    "15212345678",
					Ctime:    now,
				}).Return(nil)
				return urepo, crepo
			},

			ctx:     context.Background(),
			uid:     123,
			wantErr: nil,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(100),
				AboutMe:  "自我介绍",
				Phone:    "15212345678",
				Ctime:    now,
			},
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{
						Id:       123,
						Email:    "123@qq.com",
						Password: "123456",
						Birthday: time.UnixMilli(100),
						AboutMe:  "自我介绍",
						Phone:    "15212345678",
						Ctime:    time.UnixMilli(101),
					}, nil)
				return d, c
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(100),
				AboutMe:  "自我介绍",
				Phone:    "15212345678",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},
		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).
					Return(dao.User{}, dao.ErrRecordNotFound)
				return d, c
			},
			uid:      123,
			ctx:      context.Background(),
			wantUser: domain.User{},
			wantErr:  dao.ErrRecordNotFound,
		},
		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).
					Return(dao.User{
						Id: uid,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						AboutMe:  "自我介绍",
						Phone: sql.NullString{
							String: "15212345678",
							Valid:  true,
						},
						Ctime: 101,
						Utime: 102,
					}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: time.UnixMilli(100),
					AboutMe:  "自我介绍",
					Phone:    "15212345678",
					Ctime:    time.UnixMilli(101),
				}).Return(errors.New("redis错误"))
				return d, c
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(100),
				AboutMe:  "自我介绍",
				Phone:    "15212345678",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewCacheUserRepository(tc.mock(ctrl))
			user, err := svc.FindById(tc.ctx, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
