package main

import (
	"container/list"
)

func Equals(v1, v2 interface{}) bool {
	list1, isList1 := v1.(*list.List)
	list2, isList2 := v2.(*list.List)
	if isList1 && isList2 {
		if list1.Len() != list2.Len() {
			return false
		}
		e1 := list1.Front()
		e2 := list2.Front()
		for e1 != nil && e2 != nil {
			if !Equals(e1.Value, e2.Value) {
				return false
			}
			e1 = e1.Next()
			e2 = e2.Next()
		}
		return true
	}

	vect1, isVect1 := v1.([]interface{})
	vect2, isVect2 := v2.([]interface{})
	if isVect1 && isVect2 {
		if len(vect1) != len(vect2) {
			return false
		}
		for i := 0; i < len(vect1); i++ {
			if !Equals(vect1[i], vect2[i]) {
				return false
			}
		}
		return true
	}

	map1, isMap1 := v1.(map[interface{}]interface{})
	map2, isMap2 := v2.(map[interface{}]interface{})
	if isMap1 && isMap2 {
		for key1, val1 := range map1 {
			found := false
			for key2, val2 := range map2 {
				if Equals(key1, key2) && Equals(val1, val2) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	return v1 == v2
}
