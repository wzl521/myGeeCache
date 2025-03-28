package lru

import (
	"container/list"
)

// 缓存淘汰策略LRU: 最近最少使用，当缓存超过设定的最大值时，会移除最近最少使用的记录
// LRU 算法的实现非常简单，维护一个队列，如果某条记录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可。

type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存
	nbytes    int64                         // 当前已使用的内存
	ll        *list.List                    // 双向链表,go自带的双向链表
	cache     map[string]*list.Element      // key为string,value为双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 可选，当有记录被移除时的回调函数
}

// 双向链表节点的数据类型，在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从map中删除对应的映射
type entry struct {
	key   string
	value Value
}

// 使用Len()来计算它占用了多少字节,只要实现Len()接口的方法都属于Value类型
type Value interface {
	Len() int
}

// 实现len方法
func (c *Cache) Len() int {
	return c.ll.Len()
}

// 实例化Cache
func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 查找功能:因为被访问了，所以移动到队尾
// 1. 从字典中找到对应的双向链表的节点
// 2. 将该节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 从字典中找到对应的双向链表的节点
	if ele, ok := c.cache[key]; ok {
		// 将该节点移动到队尾,双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 删除
func (c *Cache) RemoveOldest() {
	// 取到队首节点(最近最少访问的节点)，从链表中删除
	ele := c.ll.Back() // 取到队首节点
	if ele != nil {
		// 从链表中删除该节点
		c.ll.Remove(ele)
		// 取到存储的真实值,key-value
		kv := ele.Value.(*entry)
		// 从字典中 c.cache 删除该节点的映射关系，根据map中的key删除
		delete(c.cache, kv.key)
		// 更新当前所用的内存 c.nbytes
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 如果回调函数 OnEvicted 不为 nil，则调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 新增/修改
// 需要注意的是，新增或修改可能达到内存的最大限制，从而触发删除逻辑
func (c *Cache) Add(key string, value Value) {
	// 如果键存在，则更新（修改）对应节点的值，并将该节点移到队尾
	if ele, ok := c.cache[key]; ok {
		// 存在则更新对应节点的值，并将该节点移到队尾,因为被访问了
		c.ll.MoveToFront(ele)
		// 更新值
		kv := ele.Value.(*entry)
		// 更新 c.nbytes，用传入的value的长度减去原来的长度，计算出使用的内存大小
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 不存在则是新增场景，首先队尾添加新节点 &entry{key, value}
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		// 更新 c.nbytes，新增key+value的长度
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 更新 c.nbytes，如果超过了设定的最大值 c.maxBytes，则循环移除最少访问的节点
	// 使用循环，是因为添加了一个很大的键值队，移除一次可能还不够，需要多次移除
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}
