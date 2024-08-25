package web

import (
	"errors"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            *service.UserService
}

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            userService,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// REST 风格
	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", h.SignUp)
	// POST /users/login
	ug.POST("/login", h.Login)
	// POST /users/edit
	ug.POST("/edit", h.Edit)
	// GET /users/profile
	ug.GET("/profile", h.Profile)
}
func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 校验输入
	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}
	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}
	// 调用 service 服务
	err = h.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string
		Password string
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserOrPassword) {
			ctx.String(http.StatusOK, "用户名或者密码不对")
			return
		}
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	uc := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)), // 设置过期时间
		},
		Uid:       u.Id,
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	//使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

func (h *UserHandler) Edit(ctx *gin.Context) {

}

func (h *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "成功")
}

var JWTKey = []byte("WnKX59XWgcvePtvFympqvjY2M6R5sXYw")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
