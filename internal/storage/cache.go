package storage

import (
	"container/heap"
	"fmt"
	"strings"
)

// cache在本质上是一个优先队列，依据使用次数来进行排序

// 定义存储在cache中的键值数据结构
type record struct {
	key   string
	value string
	time  int // 访问次数 用于实现LRU算法
}

// Cache 定义cache
type Cache struct {
	items []*record
	index int
	size  int
}

// Len 返回目前Cache的长度
func (c Cache) Len() int { return c.index }

// Less 用于LRU算法排序
func (c Cache) Less(i, j int) bool { return c.items[i].time < c.items[j].time }

// Swap 用于实现优先队列的交换
func (c Cache) Swap(i, j int) { c.items[i], c.items[j] = c.items[j], c.items[i] }

func (c *Cache) Push(x interface{}) {
	if c.index >= c.size {
		heap.Pop(c) // LRU算法 当缓存满的时候自动弹出最近最少使用的元素
	}
	c.items[c.index] = x.(*record)
	c.index++
}

func (c *Cache) Append(key string, value string) {
	heap.Push(c, &record{key, value, 1})
}

// Get 从cache中读取结果
func (c *Cache) Get(key string) string {
	for i := 0; i < c.index; i++ {
		if c.items[i] != nil && c.items[i].key == key {
			c.items[i].time++
			return c.items[i].value
		}
	}
	return "nil"
}

// Remove 删除指定键的元素
func (c *Cache) Remove(key string) {
	for index, ele := range c.items {
		if ele != nil && ele.key == key {
			heap.Remove(c, index)
			break
		}
	}

}

// Update 更新cache中的值
func (c *Cache) Update(key string, value string) {
	var exists bool
	for _, ele := range c.items {
		if ele != nil && ele.key == key {
			ele.time++
			ele.value = value
			exists = true
		}
	}
	if exists == false { // 在缓存中还未记录
		c.Append(key, value)
	}
}

func (c *Cache) Pop() (v interface{}) {
	v = (c.items)[c.Len()-1]
	c.index--
	return
}

// showCache 打印cache中的值
func showCache(c *Cache) string {
	var buffer strings.Builder
	fmt.Fprintf(&buffer, "The size of cache is %d\n", c.size)
	for idx, ele := range c.items {
		if ele == nil {
			fmt.Fprintf(&buffer, "Index:%d, Times:0, Key:null, Value:null\n", idx)
		} else {
			fmt.Fprintf(&buffer, "Index:%d, Times:%d, Key:%s, Value:%s\n", idx, ele.time, ele.key, ele.value)
		}
	}
	return buffer.String()
}

func NewCache(size int) *Cache {
	cache := new(Cache)
	cache.items = make([]*record, size)
	cache.size = size
	return cache
}
