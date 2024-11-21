package article

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addr = []string{"192.168.1.3:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addr, cfg)
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "article_read_event",
			Value: sarama.StringEncoder(`{"aid": 1, "uid": 123}`),
		})
	}
}
