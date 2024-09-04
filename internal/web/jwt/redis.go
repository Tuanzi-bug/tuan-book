package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	Client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

var JWTKey = []byte("WnKX59XWgcvePtvFympqvjY2M6R5sXYw")

const (
	XJWTToken     = "x-jwt-token"
	XRefreshToken = "x-refresh-token"
)

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		Client:        client,
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	// 获取传入的token值
	uc, _ := ctx.Get("user")
	claims, ok := uc.(UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return errors.New("暂无权限")
	}
	// 把token进行置空
	ctx.Set(XJWTToken, "")
	ctx.Set(XRefreshToken, "")
	// ssid 作为一个状态，存入redis，说明你执行了退出登录，设置过期时间与刷新token一致
	return h.Client.Set(ctx, h.key(claims.Ssid), claims.Ssid, h.rcExpiration).Err()
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	// 检查是否存在ssid
	cnt, err := h.Client.Exists(ctx, h.key(ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return h.SetRefreshToken(ctx, uid, ssid)
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	// 根据约定获取头部 token
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		// 没登录，没有 token, Authorization 这个头部都没有
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}
	// 按空格分割
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}
	return segs[1]
}

func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	//使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(h.signingMethod, rc)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)), // 设置过期时间
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	//使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(h.signingMethod, uc)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header(XJWTToken, tokenStr)
	return nil
}

func (h *RedisJWTHandler) key(ssid string) string {
	return fmt.Sprintf("users:ssid:%s", ssid)
}
