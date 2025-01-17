package models

import "time"

type ProductSeckill struct {
	ID         int       `gorm:"primaryKey;column:id" json:"id"`               // 主键
	Name       string    `gorm:"size:64;column:name" json:"name"`              // 活动名称
	Price      float64   `gorm:"type:decimal(11,2);column:price" json:"price"` // 活动价格
	Num        int       `gorm:"column:num" json:"num"`                        // 参与秒杀的数量
	PId        int       `gorm:"column:pid" json:"pid"`                        // 商品外键
	StartTime  time.Time `gorm:"column:start_time" json:"start_time"`          // 秒杀开始时间
	EndTime    time.Time `gorm:"column:end_time" json:"end_time"`              // 秒杀结束时间
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`        // 活动创建时间
}

// TableName 设置表名为 sys_product_seckill
func (ProductSeckill) TableName() string {
	return "sys_product_seckill"
}
