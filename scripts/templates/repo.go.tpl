package repo

import (
	"context"
	"gorm.io/gorm"
	// "tongcheng/internal/model" // 如果需要 Model，取消注释
)

// {{BizName}}Repo defines the interface for the {{bizName}} repository.
type {{BizName}}Repo interface {
	// 在这里定义 {{bizName}} 仓库的方法
	// ExampleMethod(ctx context.Context, id uint) (*model.{{BizName}}, error)
}

type {{bizName}}Repo struct {
	db *gorm.DB
	// 在这里添加其他依赖，例如 Redis 客户端
}

// New{{BizName}}Repo creates a new {{BizName}}Repo.
func New{{BizName}}Repo(db *gorm.DB) {{BizName}}Repo {
	return &{{bizName}}Repo{
		db: db,
	}
}

// 在这里实现 {{BizName}}Repo 接口的方法...
// 例如:
// func (r *{{bizName}}Repo) ExampleMethod(ctx context.Context, id uint) (*model.{{BizName}}, error) {
//     var entity model.{{BizName}}
//     err := r.db.WithContext(ctx).First(&entity, id).Error
//     if err != nil {
//         // 处理错误，例如 gorm.ErrRecordNotFound
//         return nil, err
//     }
//     return &entity, nil
// }