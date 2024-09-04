//go:build wireinject

package main

import (
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"github.com/Tuanzi-bug/tuan-book/internal/web"
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/Tuanzi-bug/tuan-book/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		// 数据层
		dao.NewUserDAO,
		// 缓存
		cache.NewUserCache,
		cache.NewCodeCache,
		// 存储层
		repository.NewCacheUserRepository,
		repository.NewCacheCodeRepository,
		// 服务层
		service.NewUserService,
		service.NewCodeService,
		ioc.InitSMSService,

		// 控制层
		web.NewUserHandler,
		myjwt.NewRedisJWTHandler,
		// 初始化服务
		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
