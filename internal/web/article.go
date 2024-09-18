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
	g.POST("/publish", h.Publish)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
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
		zap.L().Error("保存帖子失败", zap.Int64("uid", u.Uid), zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "帖子保存成功！", Data: aid})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
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
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: u.Uid},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		zap.L().Error("保存帖子失败", zap.Int64("uid", u.Uid), zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
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
	err := h.svc.Withdraw(ctx, u.Uid, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		zap.L().Error("撤回文章失败！", zap.Int64("uid", u.Uid), zap.Int64("aid", req.Id), zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}
