package integration

import (
	"bytes"
	"encoding/json"
	"github.com/Tuanzi-bug/tuan-book/internal/integration/startup"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/Tuanzi-bug/tuan-book/ioc"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ArticleTestSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

// 在所有测试执行之前，初始化一些内容
func (s *ArticleTestSuite) SetupSuite() {
	// 初始化db
	s.db = ioc.InitDB()
	hdl := startup.InitArticleHandler()
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{Uid: 123})
	})
	hdl.RegisterRoutes(server) // 注册路由
	// 模拟一些claims中间件
	s.server = server
}

// TearDownTest 每一个都会执行
func (s *ArticleTestSuite) TearDownTest() {
	// 清空所有数据，并且自增主键恢复到 1
	s.db.Exec("TRUNCATE TABLE articles")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		// 准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		//输入
		req Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {
				// 不需要做什么
			},
			after: func(t *testing.T) {
				// 验证结果
				var art dao.Article
				// 通过数据库进行验证
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				// 检查时间
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.Equal(t, "我是标题", art.Title)
				assert.Equal(t, "我是内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
			},
			req: Article{
				Title:   "我是标题",
				Content: "我是内容",
			},
			wantCode:   http.StatusOK,
			wantResult: Result[int64]{Data: 1, Msg: "帖子保存成功！"},
		},
		{
			name: "修改帖子后保存",
			before: func(t *testing.T) {
				// 修改是在已经存在的基础上进行修改,需要插入新数据
				err := s.db.Create(&dao.Article{
					Id:       11,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    456,
					Utime:    789}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证结果
				var art dao.Article
				// 通过数据库进行验证
				err := s.db.Where("id=?", 11).First(&art).Error
				assert.NoError(t, err)
				// 检查时间
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
			},
			req: Article{
				Id:      11,
				Title:   "新的标题",
				Content: "新的内容",
				Author:  Author{Id: 123},
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 11,
				Msg:  "帖子保存成功！",
			},
		},
		{
			name: "修改帖子-修改其他人的帖子",
			before: func(t *testing.T) {
				// 修改是在已经存在的基础上进行修改,需要插入新数据
				err := s.db.Create(&dao.Article{
					Id:       22,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1024,
					Ctime:    456,
					Utime:    789}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证结果
				var art dao.Article
				// 通过数据库进行验证
				err := s.db.Where("id=?", 22).First(&art).Error
				assert.NoError(t, err)
				// 检查时间
				assert.Equal(t, dao.Article{
					Id:       22,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1024,
					Ctime:    456,
					Utime:    789}, art)
			},
			req: Article{
				Id:      22,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Msg: "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			reqBody, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			// 准备Req和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit",
				bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			// 执行
			s.server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			// 对于结构体验证，进行反序列化再进行比较
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, res)
		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
}

type Author struct {
	Id int64
}

type Result[T any] struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
