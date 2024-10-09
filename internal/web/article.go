package web

import (
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
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
	g.POST("/withdraw", h.Withdraw)
	g.POST("/list", h.List)
	// 创作者相关的接口
	a := g.Group("/author")
	a.GET("/detail/:id", h.Detail)
	// 线上库的相关接口
	pub := g.Group("/pub")
	pub.GET("/detail/:id", h.PubDetail)
}

// Edit 编辑文章接口
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

// Publish 发布文章接口
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

// Withdraw 撤回文章接口
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

// List 文章列表接口
func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	uc := ctx.MustGet("user").(myjwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		zap.L().Error("查找创作者文章失败", zap.Int64("uid", uc.Uid), zap.Int("Limit", page.Limit), zap.Int("Offset", page.Offset), zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
		return ArticleVo{
			Id:       src.Id,
			Title:    src.Title,
			Abstract: src.Abstract(),

			//Content:  src.Content,
			AuthorId: src.Author.Id,
			// 列表，你不需要
			Status: src.Status.ToUint8(),
			Ctime:  src.Ctime.Format(time.DateTime),
			Utime:  src.Utime.Format(time.DateTime),
		}
	})})
}

// Detail 制作库创作者的文章详情接口
func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Msg: "id 参数错误", Code: 4})
		zap.L().Warn("Detail 获取 id 参数错误", zap.String("id", idStr), zap.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Msg: "系统错误", Code: 5})
		zap.L().Error("查找文章失败", zap.Int64("id", id), zap.Error(err))
		return
	}
	// 涉及到作者的相关信息，需要通过token信息获取个人信息
	uc := ctx.MustGet("user").(myjwt.UserClaims)
	// 校验文章作者信息与当前登录用户是否一致
	if uc.Uid != art.Author.Id {
		ctx.JSON(http.StatusOK, Result{Msg: "无权查看", Code: 4})
		zap.L().Warn("无权查看", zap.Int64("uid", uc.Uid), zap.Int64("aid", id))
		return
	}
	// 返回与前端约定的数据
	ctx.JSON(http.StatusOK, Result{Data: ArticleVo{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),

		Content:  art.Content,
		AuthorId: art.Author.Id,
		// 列表，你不需要
		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}})

}

// PubDetail 线上库文章详情接口
func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Msg: "id 参数错误", Code: 4})
		zap.L().Warn("PubDetail 获取 id 参数错误", zap.String("id", idStr), zap.Error(err))
		return
	}
	art, err := h.svc.GetPubById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
	}

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,

			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,

			Status: art.Status.ToUint8(),
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
		},
	})
}
