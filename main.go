package main

import (
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/config"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"github.com/Tuanzi-bug/tuan-book/internal/web"
	"github.com/Tuanzi-bug/tuan-book/internal/web/middleware"
	"github.com/Tuanzi-bug/tuan-book/pkg/gin-plus/middlewares/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
	server.POST("/book", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GET",
		})
	})
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
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
	})
	server.Use(ratelimit.NewBuilder(redisClient,
		time.Second, 1).Build())
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
	server.Use(login.JWTAuthMiddleware())
}
