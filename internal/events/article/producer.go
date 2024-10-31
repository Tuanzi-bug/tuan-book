package article

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

const TopicReadEvent = "article_read_event"

type ReadEvent struct {
	Aid int64
	Uid int64
}

type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

// ProduceReadEvent 生产阅读事件
func (s *SaramaSyncProducer) ProduceReadEvent(evt ReadEvent) error {
	// 序列化
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.ByteEncoder(val),
	})
	return err
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{
		producer: producer,
	}
}
