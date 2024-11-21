package ioc

import (
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
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

	// 使用gorm的prometheus插件
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "tuan_book",
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	//cb := gormx.NewCallbacks(prometheus2.SummaryOpts{
	//	Namespace: "geektime_daming",
	//	Subsystem: "webook",
	//	Name:      "gorm_db",
	//	Help:      "统计 GORM 的数据库查询",
	//	ConstLabels: map[string]string{
	//		"instance_id": "my_instance",
	//	},
	//	Objectives: map[float64]float64{
	//		0.5:   0.01,
	//		0.75:  0.01,
	//		0.9:   0.01,
	//		0.99:  0.001,
	//		0.999: 0.0001,
	//	},
	//})

	//err = db.Use(cb)
	//if err != nil {
	//	panic(err)
	//}

	err = db.AutoMigrate(&dao.User{}, &dao.Article{}, &dao.PublishedArticle{}, dao.Interactive{}, dao.UserLikeBiz{}, dao.UserCollectionBiz{})
	if err != nil {
		panic(err)
	}
	return db
}
