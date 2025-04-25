package handler

import (
	"tongcheng/internal/service" // 使用从 go.mod 获取的模块路径
	"github.com/gin-gonic/gin"
)

type {{BizName}}Handler struct {
	svc service.{{BizName}}Service
}

func New{{BizName}}Handler(svc service.{{BizName}}Service) *{{BizName}}Handler {
	return &{{BizName}}Handler{svc: svc}
}

func (h *{{BizName}}Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// 在这里注册 {{bizName}} 相关的路由
	// rg.GET("/{{bizName}}", h.Get{{BizName}})
	// rg.POST("/{{bizName}}", h.Create{{BizName}})
}

// 在这里实现具体的 handler 方法...
// 例如:
// func (h *{{BizName}}Handler) Get{{BizName}}(c *gin.Context) {
//     // 实现获取逻辑
// }
//
// func (h *{{BizName}}Handler) Create{{BizName}}(c *gin.Context) {
//     // 实现创建逻辑
// }