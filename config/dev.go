//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(192.168.1.3:3306)/tuan_book",
	},
	Redis: RedisConfig{
		Addr:     "192.168.1.3:3306",
		Password: "123456",
	},
}
