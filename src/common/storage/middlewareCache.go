package storage

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisCache struct {
	client redis.UniversalClient // 支持单节点和集群模式
}

// NewRedisClusterCache 创建 Redis cluster 缓存实例
func NewRedisClusterCache(addrs []string, password string) *RedisCache {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,
		Password: password,
	})
	return &RedisCache{
		client: client,
	}
}

// NewRedisCache 创建 Redis 单节点缓存实例
func NewRedisCache(addr string, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // 如果没有密码则传空字符串
		DB:       db,
	})
	return &RedisCache{
		client: client,
	}
}

// Get 从 Redis 缓存中获取数据
func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	result, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// 缓存中没有该键
		return nil, false, nil
	} else if err != nil {
		// 发生错误
		return nil, false, err
	}
	// 假设返回结果是字符串，你可以在此处转换为需要的类型
	return result, true, nil
}

// Set 向 Redis 缓存中设置数据
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	// 默认过期时间为 10 分钟
	if duration == 0 {
		duration = 10 * time.Minute
	}
	// 将 value 序列化为字符串
	valueStr, err := serializeValue(value)
	if err != nil {
		return err
	}

	// 设置缓存
	err = rc.client.Set(ctx, key, valueStr, duration).Err()
	return err
}

// Delete 从 Redis 缓存中删除数据
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	err := rc.client.Del(ctx, key).Err()
	return err
}

// Exists 判断 Redis 中缓存是否存在
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// 序列化 value 为字符串（支持 JSON 序列化）
func serializeValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}
