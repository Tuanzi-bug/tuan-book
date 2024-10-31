package article

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/pkg/saramax"
	"go.uber.org/zap"
	"log"
	"time"
)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
}

func NewInteractiveReadEventConsumer(client sarama.Client, repo repository.InteractiveRepository) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		repo:   repo,
	}
}

// Start 启动消费者
func (r *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	go func() {
		for {
			err := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewBatchHandler[ReadEvent](r.BatchConsume))
			if err != nil {
				// 记录日志，不影响主流程
				log.Println("consume read event failed", zap.Error(err))
			}
		}
	}()
	return err
}

// Consume 消费信息策略
func (r *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage,
	evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := r.repo.IncrReadCnt(ctx, "article", evt.Aid)
	return err
}

func (r *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}
