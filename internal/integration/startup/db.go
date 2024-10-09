package startup

import (
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg = Config{
		DSN: "root:root@tcp(192.168.1.3:3306)/tuan_book",
	}
	err := viper.UnmarshalKey("mysql", &cfg)

	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&dao.User{}, &dao.Article{}, &dao.PublishedArticle{}, dao.Interactive{})
	if err != nil {
		panic(err)
	}
	return db
}
