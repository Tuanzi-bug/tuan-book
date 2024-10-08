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

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCacheUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	dao.NewGROMArticleDAO,
	cache.NewArticleRedisCache,
	repository.NewCacheArticleRepository,
	service.NewArticleService)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,
		// 接口集合
		userSvcProvider,
		articleSvcProvider,
		// 数据层
		//dao.NewUserDAO,
		//dao.NewGROMArticleDAO,
		// 缓存
		//cache.NewUserCache,
		cache.NewCodeCache,
		//cache.NewArticleCache,
		// 存储层
		//repository.NewCacheUserRepository,
		repository.NewCacheCodeRepository,
		//repository.NewCacheArticleRepository,
		// 第三方服务层
		//service.NewUserService,
		service.NewCodeService,
		//service.NewArticleService,
		ioc.InitSMSService,

		// 控制层
		web.NewUserHandler,
		web.NewArticleHandler,
		myjwt.NewRedisJWTHandler,
		// 初始化服务
		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
