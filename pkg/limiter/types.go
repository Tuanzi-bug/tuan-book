package limiter

import "context"

// Limiter  限流器的接口
type Limiter interface {
	// Limit bool 代表是否限流，true 就是要限流 err 限流器本身有没有错误
	Limit(ctx context.Context, key string) (bool, error)
}
