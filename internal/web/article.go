package web

import (
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type ArticleHandler struct {
	svc service.ArticleService
}

func NewArticleHandler(svc service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")

	g.POST("/edit", h.Edit)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string
		Content string
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, ok := ctx.MustGet("user").(myjwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Msg: "系统错误"})
		zap.L().Error("未发现用户 session 信息")
		return
	}
	aid, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: u.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Msg: "系统错误"})
		zap.L().Error("保存帖子失败")
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "帖子保存成功！", Data: aid})
}
