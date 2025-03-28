package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

// 测试 Get 方法
func TestCache_Get(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("1234"))

	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		// 遇到错误就停止运行
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

// 测试remove方法
func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil) //20
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	// key1 应该已经被淘汰，因为缓存容量超过了 20 字节,因此，如果移除成功，则不会执行以下语句
	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatalf("RemoveOldest failed, key1 should be evicted, cache length: %d", lru.Len())
	}
}

// 测试回调函数
func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	//使用 reflect.DeepEqual 函数比较 expect 和 keys 切片是否完全相等。
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestMapDelete(t *testing.T) {
	m := make(map[string]string, 10)
	m["key1"] = "value1"
	m["key2"] = "value2"
	delete(m, "key1")
	for k, v := range m {
		t.Logf("key: %s, value: %s", k, v)
	}
}
