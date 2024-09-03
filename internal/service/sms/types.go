package sms

import "context"

type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}

// 设计并实现了一个高可用的短信平台
// 1. 提高可用性：重试机制、客户端限流、failover（轮询，实时检测）
// 	1.1 实时检测：
// 	1.1.1 基于超时的实时检测（连续超时）
// 	1.1.2 基于响应时间的实时检测（比如说，平均响应时间上升 20%）
//  1.1.3 基于长尾请求的实时检测（比如说，响应时间超过 1s 的请求占比超过了 10%）
//  1.1.4 错误率
// 2. 提高安全性：
// 	2.1 完整的资源申请与审批流程
//  2.2 鉴权：
// 	2.2.1 静态 token
//  2.2.2 动态 token
// 3. 提高可观测性：日志、metrics, tracing，丰富完善的排查手段
// 4. 提高性能，高性能：
