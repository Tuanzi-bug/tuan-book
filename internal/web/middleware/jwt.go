package middleware

import (
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type JWTMiddlewareBuilder struct {
	ignorePaths []string
	JWTHandler  myjwt.Handler
}

func NewJWTMiddlewareBuilder(hdl myjwt.Handler) *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{
		ignorePaths: make([]string, 0),
		JWTHandler:  hdl,
	}
}

func (m *JWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		for _, p := range m.ignorePaths {
			if p == path {
				return
			}
		}

		tokenStr := m.JWTHandler.ExtractToken(ctx)
		var uc myjwt.UserClaims
		// 自定义签名解析token
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return myjwt.JWTKey, nil
		})
		if err != nil {
			// token 不对，token 是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			// token 解析出来了，但是 token 可能是非法的，或者过期了的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 加入user-Agent 增加安全性
		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			// 后期我们讲到了监控告警的时候，这个地方要埋点
			// 能够进来这个分支的，大概率是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//expireTime := uc.ExpiresAt
		//// 剩余过期时间 < 50s 就要刷新
		//if expireTime.Sub(time.Now()) < time.Second*50 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 5))
		//	tokenStr, err = token.SignedString(web.JWTKey)
		//	ctx.Header("x-jwt-token", tokenStr)
		//	if err != nil {
		//		// 这边不要中断，因为仅仅是过期时间没有刷新，但是用户是登录了的
		//		log.Println(err)
		//	}
		//}
		ctx.Set("user", uc)
		ctx.Next()
	}
}

func (m *JWTMiddlewareBuilder) IgnorePaths(path string) *JWTMiddlewareBuilder {
	m.ignorePaths = append(m.ignorePaths, path)
	return m
}
