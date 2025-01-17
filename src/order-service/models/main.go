package models

import "time"

type Order struct {
	ID         int       `gorm:"primaryKey;column:id" json:"id"`            // 主键
	OrderNum   string    `gorm:"size:64;column:order_num" json:"order_num"` // 订单编号
	UId        int       `gorm:"column:uid" json:"uid"`                     // 用户外键，关联 sys_user
	SId        int       `gorm:"column:sid" json:"sid"`                     // 活动外键，关联 sys_product_seckill
	PayStatus  int       `gorm:"column:pay_status" json:"pay_status"`       // 支付状态
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`     // 订单创建时间
}

// TableName 设置表名为 sys_orders
func (Order) TableName() string {
	return "sys_orders"
}
