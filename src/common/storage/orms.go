package storage

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"reflect"
)

// ORM 数据库接口
type ORM interface {
	// 创建记录
	Create(model interface{}) error

	// 查找单条记录，根据主键查找
	Find(model interface{}, id interface{}) (interface{}, error)

	// 查找多条记录
	FindAll(model interface{}, condition FieldCondition) ([]interface{}, error)

	// 更新记录
	Update(model interface{}, id interface{}, fields map[string]interface{}) error

	// 删除记录
	Delete(model interface{}, id interface{}) error
}

type GORM struct {
	masters []*gorm.DB // 主库连接池
	slaves  []*gorm.DB // 从库连接池
}

// NewGORM 初始化 GORM 连接，支持多个主库和从库
func NewGORM(masterDSNs, slaveDSNs []string) *GORM {
	masters := make([]*gorm.DB, 0)
	slaves := make([]*gorm.DB, 0)

	// 初始化主库连接池
	for _, masterDSN := range masterDSNs {
		mysqlConn := mysql.Open(masterDSN)
		db, err := gorm.Open(mysqlConn, &gorm.Config{})
		if err != nil {
			log.Fatal("Error connecting to master DB:", err)
		}
		masters = append(masters, db)
	}

	// 初始化从库连接池
	for _, slaveDSN := range slaveDSNs {
		mysqlConn := mysql.Open(slaveDSN)
		db, err := gorm.Open(mysqlConn, &gorm.Config{})
		if err != nil {
			log.Fatal("Error connecting to slave DB:", err)
		}
		slaves = append(slaves, db)
	}

	return &GORM{masters: masters, slaves: slaves}
}
func (g *GORM) Migrate(model interface{}) error {
	g.getRandomDB(true).AutoMigrate(model)
	g.getRandomDB(false).AutoMigrate(model)
	return nil
}

// getRandomDB 根据是否写操作选择主库或从库
func (g *GORM) getRandomDB(isWrite bool) *gorm.DB {
	if isWrite {
		// 从主库中随机选择
		return g.masters[rand.Intn(len(g.masters))]
	}
	// 从从库中随机选择
	return g.slaves[rand.Intn(len(g.slaves))]
}

// Create 创建记录
func (g *GORM) Create(model interface{}) error {
	// 使用主库进行写操作
	return g.getRandomDB(true).Create(model).Error
}

// Find 查找单条记录，根据主键查找
func (g *GORM) Find(model interface{}, id interface{}) (interface{}, error) {
	// 使用从库进行读操作
	db := g.getRandomDB(true)

	// 执行查找操作
	err := db.Model(model).Where("id = ?", id).Find(model).Error
	if err != nil {
		return nil, err
	}

	// 返回查找到的记录
	return model, nil
}

// FindAll 查找多条记录，支持复杂条件查询
func (g *GORM) FindAll(model interface{}, condition FieldCondition) ([]interface{}, error) {
	// 使用从库进行读操作
	query := g.getRandomDB(false).Model(model)

	// 根据查找类型进行条件处理
	switch condition.CondType {
	case Exact:
		query = query.Where(condition.Field+" = ?", condition.Value)
	case Range:
		query = query.Where(condition.Field+" BETWEEN ? AND ?", condition.RangeMin, condition.RangeMax)
	case Like:
		query = query.Where(condition.Field+" LIKE ?", "%"+condition.Value.(string)+"%")
	}

	// 获取 model 的类型
	modelType := reflect.TypeOf(model)

	// 如果 model 是一个指针类型，我们需要获取实际的结构体类型
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 创建一个切片类型，其元素类型为 model 的类型
	sliceType := reflect.SliceOf(modelType)

	// 创建切片实例
	results := reflect.New(sliceType).Interface()

	// 使用 GORM 查询并填充结果
	err := query.Find(results).Error
	if err != nil {
		return nil, err
	}

	// 获取反射值，并转换为切片
	val := reflect.ValueOf(results).Elem()

	// 如果没有结果，返回空切片
	if val.Len() == 0 {
		return []interface{}{}, nil
	}

	// 将查询结果转换为 []interface{}
	var resultList []interface{}
	for i := 0; i < val.Len(); i++ {
		resultList = append(resultList, val.Index(i).Interface())
	}

	return resultList, nil
}

// Update 更新记录
func (g *GORM) Update(model interface{}, id interface{}, fields map[string]interface{}) error {
	// 使用主库进行写操作
	return g.getRandomDB(true).Model(model).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除记录
func (g *GORM) Delete(model interface{}, id interface{}) error {
	// 使用主库进行写操作
	return g.getRandomDB(true).Delete(model, id).Error
}
