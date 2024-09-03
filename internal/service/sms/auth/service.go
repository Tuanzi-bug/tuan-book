package auth

import (
	"context"
	"errors"
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

// 对服务加入一个权限控制，采用jwt来进行控制

type JWTDecorator struct {
	svc sms.Service
	key []byte
}

func NewJWTDecorator(svc sms.Service) *JWTDecorator {
	return &JWTDecorator{
		svc: svc,
		key: []byte("AccessControl"), // 根据调用业务方进行传入，也可以设置一个固定的
	}
}

// GenerateToken 对业务方进行签发token
func (j *JWTDecorator) GenerateToken(ctx context.Context, tplId string) (string, error) {
	claims := SMSClaims{
		Tpl: tplId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.key)
}

func (j *JWTDecorator) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims
	// 解析token，能解析成功，说明是我们签发的。
	token, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return j.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("校验不通过")
	}
	return j.Send(ctx, claims.Tpl, args, numbers...)
}

type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
}
