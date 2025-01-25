package storage

// ORM 数据库接口
type ORM interface {
	// 创建记录
	Create(model interface{}) error

	// 查找单条记录，根据主键查找
	Find(model interface{}, id interface{}) error

	// 查找多条记录
	FindAll(model interface{}, conditions map[string]interface{}) ([]interface{}, error)

	// 更新记录
	Update(model interface{}, id interface{}, fields map[string]interface{}) error

	// 删除记录
	Delete(model interface{}, id interface{}) error
}
