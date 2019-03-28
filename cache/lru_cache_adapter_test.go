package cache

import (
	"reflect"
	"testing"
)

func Test_LRU(t *testing.T)  {
	var v interface{}
	cache := newLRUCacheAdapter(5, 3)

	cache.put("k1", "hello1", 0)
	cache.put("k2", "hello2", 0)
	cache.put("k3", "hello3", 0)
	cache.put("k4", "hello4", 0)
	cache.put("k5", "hello5", 0)

	if k1 := cache.get("k1"); k1 != nil {
		t.Error("get k1:", cache.get("k1"))
	}
	if k1 := cache.get("k2"); k1 != nil {
		t.Error("get k2:", cache.get("k2"))
	}
	/*******************************************************************/
	cache.del("k3")
	v = nil
	if reflect.DeepEqual(v, cache.get("k3")) == false {
		t.Error("get k3:", cache.get("k3"))
	}
	/*******************************************************************/
	cache.clear()
	v = nil
	if reflect.DeepEqual(v, cache.get("k4")) == false {
		t.Error("get k4:", cache.get("k4"))
	}


}
