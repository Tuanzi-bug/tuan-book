package ioc

import (
	"github.com/IBM/sarama"
	"github.com/Tuanzi-bug/tuan-book/internal/events"
	"github.com/Tuanzi-bug/tuan-book/internal/events/article"
	"github.com/spf13/viper"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `json:"addr"`
	}
	// 从配置中读取kafka的配置
	var cfg Config
	// 读取配置
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addr, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(c1 *article.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
