package tongchengapi

// Options 定义灵活的可选参数
type Options map[string]interface{}


// 创建新的Options实例的工厂函数
func NewOptions() Options {
    return make(Options)
}

// 设置值并返回Options，支持链式调用
func (o Options) Set(key string, value interface{}) Options {
    o[key] = value
    return o
}
// 删除键并返回Options，支持链式调用
func (o Options) Delete(key string) Options {
    delete(o, key)
    return o
}
// 获取值，不存在则返回nil
func (o Options) Get(key string) interface{} {
    return o[key]
}
// 检查键是否存在
func (o Options) Has(key string) bool {
    _, exists := o[key]
    return exists
}
// 合并另一个Options，返回新Options
func (o Options) Merge(other Options) Options {
    result := make(Options)
    for k, v := range o {
        result[k] = v
    }
    for k, v := range other {
        result[k] = v
    }
    return result
}
// 克隆当前Options
func (o Options) Clone() Options {
    result := make(Options)
    for k, v := range o {
        result[k] = v
    }
    return result
}