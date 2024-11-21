package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	svcmocks "github.com/Tuanzi-bug/tuan-book/internal/service/mocks"
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				articleSvc := svcmocks.NewMockArticleService(ctrl)
				articleSvc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return articleSvc
			},
			reqBody: `
			{
				"title": "我的标题",
				"content": "我的内容"
			}`,
			wantCode: http.StatusOK,
			wantRes:  Result{Data: float64(1)},
		},
		{
			name: "修改并且发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
"id": 1,
 "title": "新的标题",
 "content": "新的内容"
}
`,
			wantCode: 200,
			wantRes: Result{
				// 原本是 int64的，但是因为 Data 是any，所以在反序列化的时候，
				// 用的 float64
				Data: float64(1),
			},
		},
		{
			name: "输入有误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
			},
			reqBody: `
{
"id": 1,
 "title": "新的标题",
 "content": "新的内容",,,,
}
`,
			wantCode: 400,
		},
		{
			name: "publish错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("mock error"))
				return svc
			},
			reqBody: `
{
"id": 1,
 "title": "新的标题",
 "content": "新的内容"
}
`,
			wantCode: 200,
			wantRes: Result{
				// 原本是 int64的，但是因为 Data 是any，所以在反序列化的时候，
				// 用的 float64
				Msg:  "系统错误",
				Code: 5,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 启动mock控制器
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h := NewArticleHandler(tc.mock(ctrl), nil)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 准备服务器，注册路由
			server := gin.Default()
			server.Use(func(context *gin.Context) {
				context.Set("user", myjwt.UserClaims{Uid: 123})
			})
			h.RegisterRoutes(server)

			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			// 对结果进行反序列化
			res := Result{}
			err = json.NewDecoder(resp.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
