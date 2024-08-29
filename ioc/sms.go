package ioc

import (
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms"
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 对SMS服务进行更换
	return memory.NewService()
}
