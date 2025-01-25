package storage

import (
	"context"
	"errors"
	"fmt"
	"gid"
	"log"
	"strings"
	"time"
)

// ConditionType 用于表示查找的条件类型
type ConditionType int

const (
	Exact ConditionType = iota // 精确查找
	Range                      // 范围查找
	Like                       // 模糊查找
)

// FieldCondition 用于描述某个字段的查找条件
type FieldCondition struct {
	Field    string        // 字段名
	Value    interface{}   // 字段的值
	CondType ConditionType // 查找类型：精确、范围、模糊等
	RangeMin interface{}   // 范围查找下限
	RangeMax interface{}   // 范围查找上限
}

type STData struct {
	value interface{}
	gid   gid.GID
}

// Storage 接口定义
type Storage interface {

	// 设置数据
	Storage(ctx context.Context, data *STData) error

	// 更新数据
	Update(ctx context.Context, old *STData, new *STData) error

	// 删除数据
	Delete(ctx context.Context, data *STData) error

	// 条件查询（包含精确查找和范围查找，支持链式）
	Filter(ctx context.Context, model *STData, condition FieldCondition) *QuerySet
}

type QuerySet struct {
	res []interface{}
}

func (s *QuerySet) Filter(model *STData, conditions FieldCondition) *QuerySet {
	var result []interface{}
	for _, item := range s.res {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			// 如果item无法转换为map[string]interface{}，则跳过
			continue
		}

		// 根据条件类型进行判断
		switch conditions.CondType {
		case Exact:
			// 精确查找
			if value, found := itemMap[conditions.Field]; found {
				if value == conditions.Value {
					result = append(result, item)
				}
			}

		case Range:
			// 范围查找
			if value, found := itemMap[conditions.Field]; found {
				switch v := value.(type) {
				case int, int32, int64, float32, float64:
					// 数字范围查找
					minValue, minOk := conditions.RangeMin.(float64)
					maxValue, maxOk := conditions.RangeMax.(float64)
					if minOk && maxOk {
						// 强制转换为 float64 类型
						if numValue, ok := v.(float64); ok {
							if numValue >= minValue && numValue <= maxValue {
								result = append(result, item)
							}
						}
					} else {
						// 错误处理：范围值类型不匹配
						continue
					}

				case time.Time:
					// 时间范围查找
					minTime, minOk := conditions.RangeMin.(time.Time)
					maxTime, maxOk := conditions.RangeMax.(time.Time)
					if minOk && maxOk {
						// 时间比较
						if v.After(minTime) && v.Before(maxTime) {
							result = append(result, item)
						}
					} else {
						// 错误处理：时间范围值类型不匹配
						continue
					}

				default:
					// 如果值不是我们期望的类型，则跳过
					continue
				}
			}

		case Like:
			// 模糊查找
			if value, found := itemMap[conditions.Field].(string); found {
				if strValue, ok := conditions.Value.(string); ok && strings.Contains(value, strValue) {
					result = append(result, item)
				}
			}

		default:
			// 未知的条件类型
			continue
		}
	}
	return &QuerySet{res: result}
}

// And 求两个 QuerySet 的交集
func (s *QuerySet) And(set *QuerySet) *QuerySet {
	// 创建一个新的 QuerySet，保存两个查询集的交集条件
	intersection := []interface{}{}

	// 使用 map 来查找交集条件
	setMap := make(map[interface{}]bool)
	for _, condition := range set.res {
		setMap[condition] = true
	}

	// 遍历当前的 QuerySet，检查条件是否在另一个 QuerySet 中出现
	for _, condition := range s.res {
		if _, exists := setMap[condition]; exists {
			intersection = append(intersection, condition)
		}
	}

	// 返回交集查询集
	return &QuerySet{
		res: intersection,
	}
}

// Or 求两个 QuerySet 的并集
func (s *QuerySet) Or(set *QuerySet) *QuerySet {
	// 创建一个新的 QuerySet，保存两个查询集的并集条件
	union := append(s.res, set.res...) // 将两个 res 合并

	// 使用 map 去重重复的条件
	setMap := make(map[interface{}]bool)
	result := []interface{}{}

	for _, condition := range union {
		if _, exists := setMap[condition]; !exists {
			result = append(result, condition)
			setMap[condition] = true
		}
	}

	// 返回并集查询集
	return &QuerySet{
		res: result,
	}
}
func (s *QuerySet) First() interface{} {
	return s.res[0]
}
func (s *QuerySet) Last() interface{} {
	return s.res[len(s.res)-1]
}
func (s *QuerySet) GetByIndex(index int) interface{} {
	return s.res[index]
}
func (s *QuerySet) GetAll() []interface{} {
	return s.res
}
func (s *QuerySet) Count() int {
	return len(s.res)
}

type BaseStorage[T Key] struct {
	LocalCache      Cache[T]
	MiddlewareCache Cache[T]
	ORM             ORM
}

// 构造函数，初始化 BaseStorage
func NewBaseStorage[T Key](localCache Cache[T], middlewareCache Cache[T], orm ORM) *BaseStorage[T] {
	return &BaseStorage[T]{
		LocalCache:      localCache,
		MiddlewareCache: middlewareCache,
		ORM:             orm,
	}
}

// 获取数据，依次从本地缓存、缓存中间件和 ORM 中查找
func (s *BaseStorage[T]) Filter(ctx context.Context, model *STData, condition FieldCondition) *QuerySet {
	// 先从本地缓存获取
	if value, found, err := s.LocalCache.Get(ctx, key); found {
		log.Println("Found in local cache")
		return value, nil
	}

	// 再从缓存中间件获取
	if value, found, err := s.MiddlewareCache.Get(ctx, key); found {
		log.Println("Found in middleware cache")
		s.LocalCache.Set(ctx, key, value, 3600) // 放到本地缓存中
		return value, nil
	}

	// 最后从 ORM 中获取
	if value, found := s.ORM.Get(key); found {
		log.Println("Found in ORM")
		s.LocalCache.Set(ctx, key, value, 3600)      // 放到本地缓存中
		s.MiddlewareCache.Set(ctx, key, value, 3600) // 放到缓存中间件
		return value, nil
	}

	return nil, nil
}

// 设置数据
func (s *BaseStorage[T]) Storage(ctx context.Context, data *STData) error {
	// 存储操作使用orm直接存储即可
	err := s.ORM.Create(value)
	return err
}

// 更新数据
func (s *BaseStorage[T]) Update(ctx context.Context, old *STData, new *STData) error {
	// 更新数据时策略：
	//1 命中本地缓存刷新缓存时间，失效时（换出或者失效）写入下级存储系统；未命中去下级缓存更新
	//2 命中时清楚缓存项，发送更新请求到消息队列，下级缓存系统从队列拿到更新消息后更新；未命中去下级更新** [采用]
	var GID T
	switch any(GID).(type) {
	case string: // 假设 base64 是字符串类型
		item, _ := old.gid.GetBase64()
		GID = any(item).(T)
	case int64:
		item, _ := old.gid.GetInt64()
		GID = any(item).(T)
	default:
		return errors.New("invalid gid type")
	}
	ex, err := s.LocalCache.Exists(ctx, GID)
	if err != nil {
		return err
	}
	if ex {
		s.LocalCache.Delete(ctx, GID)
	}
	ex, err = s.MiddlewareCache.Exists(ctx, GID)
	if err != nil {
		return err
	}
	if ex {
		s.MiddlewareCache.Delete(ctx, GID)
	}
	// todo send update msg to mq
	//err = s.ORM.Update(old.value,old.gid,new.value.(map[string]interface{}))
	if err != nil {
		return err
	}
	return nil
}

// 删除数据
func (s *BaseStorage[T]) Delete(ctx context.Context, data *STData) error {
	var GID T
	switch any(GID).(type) {
	case string: // 假设 base64 是字符串类型
		item, _ := data.gid.GetBase64()
		GID = any(item).(T)
	case int64:
		item, _ := data.gid.GetInt64()
		GID = any(item).(T)
	default:
		return errors.New("invalid gid type")
	}
	// 调用 LocalCache 的删除方法
	if err := s.LocalCache.Delete(ctx, GID); err != nil {
		return fmt.Errorf("failed to delete from local cache: %w", err)
	}

	return nil
}
