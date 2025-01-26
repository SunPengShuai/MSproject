package test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"product-service/models"
	"storage"
	"testing"
	"time"
)

func TestRedisCache(t *testing.T) {
	// 创建 Redis 缓存实例，这里使用默认的 localhost Redis 实例
	//addrs := []string{"172.16.1.1:6379", "172.16.1.2:6379", "172.16.1.3:6379", "172.16.1.4:6379", "172.16.1.5:6379", "172.16.1.6:6379"}
	//cache := storage.NewRedisClusterCache(addrs, "")
	cache := storage.NewRedisCache("localhost:6379", "123456", 0)
	// 测试 Set 和 Get
	t.Run("Test Set and Get", func(t *testing.T) {
		key := "test_key"
		value := models.Product{
			Name:  "testRedis",
			Price: 123.2,
			Num:   123,
		}

		// 将数据设置到缓存
		err := cache.Set(context.Background(), key, value, 5*time.Second)
		assert.NoError(t, err)

		// 从缓存中获取数据
		result, found, err := cache.Get(context.Background(), key)
		obj := models.Product{}
		json.Unmarshal([]byte(result.(string)), &obj)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, value, obj)
	})

	// 测试不存在的缓存
	t.Run("Test Get Non-Existent", func(t *testing.T) {
		key := "non_existent_key"
		result, found, err := cache.Get(context.Background(), key)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, result)
	})

	// 测试 Delete
	t.Run("Test Delete", func(t *testing.T) {
		key := "test_delete_key"
		value := "delete_value"

		// 设置一个键值对
		err := cache.Set(context.Background(), key, value, 5*time.Second)
		assert.NoError(t, err)

		// 删除该键值对
		err = cache.Delete(context.Background(), key)
		assert.NoError(t, err)

		// 再次尝试获取该键
		result, found, err := cache.Get(context.Background(), key)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, result)
	})

	// 测试 Exists
	t.Run("Test Exists", func(t *testing.T) {
		key := "test_exists_key"
		value := "exists_value"

		// 设置键值对
		err := cache.Set(context.Background(), key, value, 5*time.Second)
		assert.NoError(t, err)

		// 检查该键是否存在
		exists, err := cache.Exists(context.Background(), key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// 删除该键
		err = cache.Delete(context.Background(), key)
		assert.NoError(t, err)

		// 再次检查该键是否存在
		exists, err = cache.Exists(context.Background(), key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
