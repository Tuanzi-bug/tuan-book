package failover

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms"
	"sync/atomic"
)

type TimeoutFailoverDecorator struct {
	svcs []sms.Service
	// 当前正在使用节点
	idx int32
	// 连续几个超时了
	cnt int32
	// 切换的阈值，只读的
	threshold int32
	// 根据当前服务商超时的个数，达到阈值后进行切换
}

func NewTimeoutFailoverDecorator(svcs []sms.Service, threshold int32) *TimeoutFailoverDecorator {
	return &TimeoutFailoverDecorator{
		svcs:      svcs,
		threshold: threshold,
	}
}

func (t *TimeoutFailoverDecorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 获取当前下标
	idx := atomic.LoadInt32(&t.idx)
	// 记录超时个数
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt >= t.threshold {
		newidx := (idx + 1) % int32(len(t.svcs))
		// 重置超时记录
		if atomic.CompareAndSwapInt32(&t.idx, idx, newidx) {
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newidx
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	// 根据服务状态进行处理
	switch err {
	case nil:
		// 没有问题
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
		return err
	default:
		// 在这里你可以根据实际情况决定如何处理
		// 你可以考虑，换下一个，语义则是：
		// - 超时错误，可能是偶发的，我尽量再试试
		// - 非超时，我直接下一个
		// 在这里考虑直接切换
		atomic.StoreInt32(&t.cnt, 0)
		atomic.StoreInt32(&t.idx, (idx+1)%int32(len(t.svcs)))
		return err
	}
}
