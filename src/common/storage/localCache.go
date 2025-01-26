package storage

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"
)

type LocalCache[T Key] struct {
	capacity int                 // 缓存容量
	data     map[T]*list.Element // 存储缓存项
	lruList  *list.List          // 双向链表用于实现 LRU
	lock     sync.RWMutex        // 读写锁，确保线程安全
	ticker   *time.Ticker        // 定时器按时清理过期数据
	stopChan chan struct{}       // 用于停止清理协程
	stopWG   sync.WaitGroup      // 等待清理协程停止
	shutdown bool                // 缓存是否关闭
}

type cacheItem[T Key] struct {
	key        T
	value      interface{}
	expiration time.Time // 过期时间
}

// NewCache 创建新的缓存
func NewCache[T Key](capacity int, cleanTime time.Duration) *LocalCache[T] {
	cache := &LocalCache[T]{
		capacity: capacity,
		data:     make(map[T]*list.Element),
		lruList:  list.New(),
		ticker:   time.NewTicker(cleanTime),
		stopChan: make(chan struct{}),
	}

	// 启动定期清理的后台协程
	cache.stopWG.Add(1)
	go cache.cleanupExpiredItems()

	return cache
}

// Get 从缓存中获取数据
func (c *LocalCache[T]) Get(ctx context.Context, key T) (interface{}, bool, error) {
	if c.shutdown {
		return nil, false, errors.New("local cache shutting down")
	}
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.shutdown {
		return nil, false, nil
	}

	if elem, found := c.data[key]; found {
		// 检查是否过期
		item := elem.Value.(*cacheItem[T])
		if time.Now().Before(item.expiration) {
			// 移动到链表头部 (最近使用)
			c.lruList.MoveToFront(elem)
			return item.value, true, nil
		}
		// 如果过期，删除缓存项
		c.removeElement(elem)
	}
	return nil, false, nil
}

// Set 向缓存中添加或更新数据
func (c *LocalCache[T]) Set(ctx context.Context, key T, value interface{}, ttl time.Duration) error {
	if c.shutdown {
		return errors.New("local cache shutting down")
	}
	c.lock.Lock()
	defer c.lock.Unlock()

	// 如果已经存在，更新并移动到链表头部
	if elem, found := c.data[key]; found {
		elem.Value.(*cacheItem[T]).value = value
		elem.Value.(*cacheItem[T]).expiration = time.Now().Add(ttl)
		c.lruList.MoveToFront(elem)
	} else {
		// 否则插入新条目
		if c.lruList.Len() == c.capacity {
			// 删除最旧的条目（LRU）
			oldest := c.lruList.Back()
			if oldest != nil {
				c.removeElement(oldest)
			}
		}
		newItem := &cacheItem[T]{
			key:        key,
			value:      value,
			expiration: time.Now().Add(ttl),
		}
		newElem := c.lruList.PushFront(newItem)
		c.data[key] = newElem
	}
	return nil
}

// Exists 判断缓存是否存在
func (c *LocalCache[T]) Exists(ctx context.Context, key T) (bool, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, found := c.data[key]
	return found, nil
}

// Delete 删除缓存元素
func (c *LocalCache[T]) Delete(ctx context.Context, key T) error {
	c.removeElement(c.data[key])
	return nil
}

// removeElement 从缓存中删除元素
func (c *LocalCache[T]) removeElement(elem *list.Element) {
	c.lruList.Remove(elem)
	delete(c.data, elem.Value.(*cacheItem[T]).key)
}

// CleanUp 清理过期缓存项
func (c *LocalCache[T]) CleanUp() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for e := c.lruList.Back(); e != nil; {
		next := e.Prev()
		item := e.Value.(*cacheItem[T])
		if time.Now().After(item.expiration) {
			c.removeElement(e)
		}
		e = next
	}
}

// cleanupExpiredItems 启动后台定时清理协程
func (c *LocalCache[T]) cleanupExpiredItems() {
	defer c.stopWG.Done()

	for {
		select {
		case <-c.ticker.C:
			c.CleanUp()
		case <-c.stopChan:
			// 停止清理协程
			c.ticker.Stop()
			return
		}
	}
}

// Stop 清理并停止定时器和清理协程
func (c *LocalCache[T]) Stop() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.shutdown {
		return
	}

	c.shutdown = true
	// 发送停止信号，优雅地停止清理协程
	close(c.stopChan)
	// 等待清理协程退出
	c.stopWG.Wait()
}
