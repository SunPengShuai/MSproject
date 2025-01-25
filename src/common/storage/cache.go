package storage

import (
	"context"
	"time"
)

// 缓存的基础接口定义
type Key interface {
	~string | ~int64 | ~int32 | ~int
}

type Cache[T Key] interface {
	// 获取缓存数据
	Get(ctx context.Context, key T) (interface{}, bool, error)

	// 设置缓存数据
	Set(ctx context.Context, key T, value interface{}, expiration time.Duration) error

	// 删除缓存数据
	Delete(ctx context.Context, key T) error

	// 检查缓存是否存在
	Exists(ctx context.Context, key T) (bool, error)
}
