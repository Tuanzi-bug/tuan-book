package failover

import (
	"context"
	"errors"
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms"
	"log"
	"sync/atomic"
)

// 采用轮询的方式进行自动切换服务商

type FailoverDecorator struct {
	svcs []sms.Service

	// v1 的字段
	// 当前服务商下标
	idx uint64
}

func NewFailoverDecorator(svcs []sms.Service) *FailoverDecorator {
	return &FailoverDecorator{
		svcs: svcs,
	}
}

func (f *FailoverDecorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 遍历每一个服务商，直至有一个正常的服务商发送成功
	for _, s := range f.svcs {
		err := s.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
	/*
		该方法有几个缺点
		1. 每次都是从第一个服务商开始的，可以轮询下标（防止服务商压力太大，需要分布均匀）
		2. 没有处理超时、被取消情况
	*/
}

// SendV1 基于第一版进行优化，包含下标的轮询、处理超时、被取消情况。
func (f *FailoverDecorator) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 新的下标
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	// 轮询所有服务商
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			// 被取消、超时
			return err
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}
