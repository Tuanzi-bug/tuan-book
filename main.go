package main

import (
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/config"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms/memory"
	"github.com/Tuanzi-bug/tuan-book/internal/web"
	"github.com/Tuanzi-bug/tuan-book/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	ud := dao.NewUserDAO(db)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
	})
	ur := repository.NewUserRepository(ud, cache.NewUserCache(redisClient))
	us := service.NewUserService(ur)
	uc := repository.NewCodeRepository(cache.NewCodeCache(redisClient))
	ucode := service.NewCodeService(uc, memory.NewService())
	hdl := web.NewUserHandler(us, ucode)
	hdl.RegisterRoutes(server)
	_ = server.Run(":8080")
}
func initWebServer() *gin.Engine {
	server := gin.Default()
	// 跨域
	server.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"x-jwt-token"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			fmt.Println(origin)
			return true
		},
		//MaxAge: 12 * time.Hour,
	}))
	// 限流
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr:     config.Config.Redis.Addr,
	//	Password: config.Config.Redis.Password,
	//})
	//server.Use(ratelimit.NewBuilder(redisClient,
	//	time.Second, 1).Build())
	useJWT(server)
	return server
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&dao.User{})
	if err != nil {
		panic(err)
	}
	return db
}
func useJWT(server *gin.Engine) {
	login := middleware.JWTMiddlewareBuilder{}
	server.Use(login.IgnorePath("/users/signup").
		IgnorePath("/users/login").
		IgnorePath("/users/login_sms/code/send").
		IgnorePath("/users/login/sms").
		JWTAuthMiddleware())
}
