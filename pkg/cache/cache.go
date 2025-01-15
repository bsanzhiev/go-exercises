package cache

import (
	"sync"
	"time"
)

/*
Конкурентный кэш с временем жизни элементов:
Создайте потокобезопасный кэш с поддержкой конкурентного доступа.
Реализуйте автоматическое удаление устаревших элементов, используя горутины и таймеры.
Обеспечьте корректную работу при одновременном чтении, записи и удалении элементов.
*/

/*
Условие:
Реализовать потокобезопасный кэш с поддержкой конкурентного доступа и временем жизни элементов.
Требования:
- Автоматическое удаление устаревших элементов
- Потокобезопасные операции чтения/записи
- Возможность установки TTL для каждого элемента
- Периодическая очистка устаревших элементов
*/

type CacheItem struct {
	Value    interface{}
	ExpireAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	cleanup time.Duration
	items   map[string]CacheItem
}

func NewCache(cleanup time.Duration) *Cache {
	cache := &Cache{
		items:   make(map[string]CacheItem),
		cleanup: cleanup,
	}
	go cache.startCleanupTimer()
	return cache
}

func (c *Cache) startCleanupTimer() {
	ticker := time.NewTicker(c.cleanup)
	for range ticker.C {
		c.mu.Lock()
		for key, item := range c.items {
			if time.Now().After(item.ExpireAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		Value:    value,
		ExpireAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpireAt) {
		return nil, false
	}

	return item.Value, true
}
