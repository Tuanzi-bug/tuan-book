package main

import (
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"github.com/Tuanzi-bug/tuan-book/internal/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
		ExposeHeaders:    []string{},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			fmt.Println(origin)
			return true
		},
		//MaxAge: 12 * time.Hour,
	}))
	return server
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(192.168.1.3:3306)/tuan_book"))
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&dao.User{})
	if err != nil {
		panic(err)
	}
	return db
}
