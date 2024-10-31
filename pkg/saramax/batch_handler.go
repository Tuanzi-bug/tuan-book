package saramax

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"time"
)

type BatchHandler[T any] struct {
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
}

func NewBatchHandler[T any](fn func(msgs []*sarama.ConsumerMessage, ts []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{fn: fn}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10
	for {
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var done = false
		for i := 0; i < batchSize && !done; i++ {
			select {
			// 超时情况
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				//batch = append(batch, msg)
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					zap.L().Error("反序列化失败", zap.String("topic", msg.Topic), zap.Int32("partition", msg.Partition), zap.Int64("offset", msg.Offset), zap.Error(err))
					continue
				}
				batch = append(batch, msg)
				ts = append(ts, t)
			}
		}
		// 凑够一批信息时候，进行处理
		cancel()
		err := b.fn(batch, ts)
		if err != nil {
			zap.L().Error("处理消息失败", zap.Error(err))
		}
		// 标记消息
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}
