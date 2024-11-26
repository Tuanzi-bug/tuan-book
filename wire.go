//go:build wireinject

package main

import (
	"github.com/Tuanzi-bug/tuan-book/internal/events/article"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"github.com/Tuanzi-bug/tuan-book/internal/web"
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/Tuanzi-bug/tuan-book/ioc"
	"github.com/google/wire"
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCacheUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	dao.NewGORMArticleDAO,
	cache.NewArticleRedisCache,
	repository.NewCacheArticleRepository,
	article.NewSaramaSyncProducer,
	article.NewInteractiveReadEventConsumer,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB,
		ioc.InitRedis,
		//ioc.InitLogger,
		ioc.InitSyncProducer,
		ioc.InitSaramaClient,
		ioc.InitConsumers,
		// 接口集合
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcSet,
		rankingSvcSet,

		// 定时任务
		ioc.InitJobs,
		ioc.InitRankingJob,
		// 数据层
		//dao.NewUserDAO,
		//dao.NewGORMArticleDAO,
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
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
