package models

import "time"

type Product struct {
	ID         int       `gorm:"primaryKey;column:id" json:"id"`               // 主键
	Name       string    `gorm:"size:64;column:name" json:"name"`              // 商品名称
	Price      float64   `gorm:"type:decimal(11,2);column:price" json:"price"` // 价格，保留两位小数
	Num        int       `gorm:"column:num" json:"num"`                        // 商品数量
	Unit       string    `gorm:"size:32;column:unit" json:"unit"`              // 商品单位
	Pic        string    `gorm:"size:255;column:pic" json:"pic"`               // 商品图片
	Desc       string    `gorm:"size:255;column:desc" json:"desc"`             // 商品描述
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`        // 用户创建时间
}

// TableName 设置表名为 sys_product
func (Product) TableName() string {
	return "sys_product"
}
