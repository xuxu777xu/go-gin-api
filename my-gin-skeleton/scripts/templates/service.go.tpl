package service

import (
	"context"
	"tongcheng/internal/repo" // 使用从 go.mod 获取的模块路径
	// "tongcheng/internal/dto" // 如果需要 DTO，取消注释
)

// {{BizName}}Service defines the interface for the {{bizName}} service.
type {{BizName}}Service interface {
	// 在这里定义 {{bizName}} 服务的方法
	// ExampleMethod(ctx context.Context, req *dto.{{BizName}}Request) (*dto.{{BizName}}Response, error)
}

type {{bizName}}Service struct {
	repo repo.{{BizName}}Repo
	// 在这里添加其他依赖，例如其他 service
}

// New{{BizName}}Service creates a new {{BizName}}Service.
func New{{BizName}}Service(repo repo.{{BizName}}Repo) {{BizName}}Service {
	return &{{bizName}}Service{
		repo: repo,
	}
}

// 在这里实现 {{BizName}}Service 接口的方法...
// 例如:
// func (s *{{bizName}}Service) ExampleMethod(ctx context.Context, req *dto.{{BizName}}Request) (*dto.{{BizName}}Response, error) {
//     // 实现业务逻辑
//     return &dto.{{BizName}}Response{}, nil
// }