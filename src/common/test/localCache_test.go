package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"storage"
	"testing"
	"time"
)

func TestLocalCache(t *testing.T) {
	// 创建一个容量为 3 的缓存，清理时间为 1 秒
	cache := storage.NewCache[string](3, time.Second)

	// 设置测试数据
	ctx := context.Background()

	// 测试 Set 和 Get 操作
	cache.Set(ctx, "key1", "value1", 2*time.Second)
	cache.Set(ctx, "key2", "value2", 2*time.Second)
	cache.Set(ctx, "key3", "value3", 2*time.Second)

	// 获取数据并验证
	val, found, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "value1", val)

	// 测试容量限制：插入超过缓存容量的数据，最旧的键应该被移除
	cache.Set(ctx, "key4", "value4", 2*time.Second)

	val, found, err = cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.False(t, found) // 因为 key2 是最旧的，被移除了
	// 测试 Exists 操作
	exists, err := cache.Exists(ctx, "key3")
	assert.NoError(t, err)
	assert.True(t, exists) // key3 存在
	// 测试过期项：等待超过 TTL，key1 应该过期
	time.Sleep(3 * time.Second)

	val, found, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.False(t, found) // key1 已经过期

	// 测试 Delete 操作
	cache.Set(ctx, "key5", "value5", 2*time.Second)
	err = cache.Delete(ctx, "key5")
	assert.NoError(t, err)

	val, found, err = cache.Get(ctx, "key5")
	assert.NoError(t, err)
	assert.False(t, found) // key5 已被删除

	// 测试 Exists 操作
	exists, err = cache.Exists(ctx, "key3")
	assert.NoError(t, err)
	assert.False(t, exists) // key3 存在

	// 测试清理过期项
	cache.Set(ctx, "key6", "value6", 1*time.Second)
	time.Sleep(2 * time.Second) // 等待 key6 过期
	cache.CleanUp()

	val, found, err = cache.Get(ctx, "key6")
	assert.NoError(t, err)
	assert.False(t, found) // key6 已过期并被清理
}

func TestCacheStop(t *testing.T) {
	// 创建并停止缓存
	cache := storage.NewCache[string](3, time.Second)
	cache.Stop()

	// 尝试在已停止的缓存中进行操作
	err := cache.Set(context.Background(), "key1", "value1", 2*time.Second)
	assert.Error(t, err)

	val, found, err := cache.Get(context.Background(), "key1")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Equal(t, nil, val)

	// 缓存已关闭后不允许进行其他操作，执行任何写操作应该返回 nil
	cache.Stop()
	err = cache.Set(context.Background(), "key2", "value2", 2*time.Second)
	assert.Error(t, err) // 不会报错，但不会写入

	val, found, err = cache.Get(context.Background(), "key2")
	assert.Error(t, err)
	assert.False(t, found) // key2 不会被写入
}
