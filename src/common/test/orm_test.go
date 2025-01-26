package test

import (
	"log"
	"storage"
	"testing"
)

// 示例的 User 模型
type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
	Age  int
}

// 示例的 main 函数
func TestOrm(t *testing.T) {
	// 初始化 GORM 连接
	masterDSNs := []string{
		"root:root@tcp(localhost:3307)/ms?charset=utf8mb4&parseTime=True&loc=Local",
	}
	slaveDSNs := []string{
		"root:root@tcp(localhost:3308)/ms?charset=utf8mb4&parseTime=True&loc=Local",
	}

	// 创建 GORM 实例
	db := storage.NewGORM(masterDSNs, slaveDSNs)
	db.Migrate(User{})
	// 创建记录
	err := db.Create(&User{Name: "Alice", Age: 30})
	if err != nil {
		log.Fatal("Error creating record:", err)
	}

	// 查找单条记录
	var user User
	res, err := db.Find(&user, 7)
	if err != nil {
		log.Fatal("Error finding record:", err)
	}
	log.Println(res)
	// 查找多条记录，使用精确查找
	users, err := db.FindAll(&User{}, storage.FieldCondition{Field: "age", Value: 30, CondType: storage.Exact})
	if err != nil {
		log.Fatal("Error finding all records:", err)
	}

	// 查找多条记录，使用范围查找
	usersInAgeRange, err := db.FindAll(&User{}, storage.FieldCondition{Field: "age", RangeMin: 20, RangeMax: 40, CondType: storage.Range})
	if err != nil {
		log.Fatal("Error finding all records:", err)
	}

	// 查找多条记录，使用模糊查找
	usersWithNameLike, err := db.FindAll(&User{}, storage.FieldCondition{Field: "name", Value: "Ali", CondType: storage.Like})
	if err != nil {
		log.Fatal("Error finding all records:", err)
	}

	// 更新记录
	err = db.Update(&User{}, 7, map[string]interface{}{"Age": 31})
	if err != nil {
		log.Fatal("Error updating record:", err)
	}

	// 删除记录
	err = db.Delete(&User{}, 8)
	if err != nil {
		log.Fatal("Error deleting record:", err)
	}

	// 打印查询结果
	log.Println("Users found:", users)
	log.Println("Users in age range:", usersInAgeRange)
	log.Println("Users with name like 'Ali':", usersWithNameLike)
}
