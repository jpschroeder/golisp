package main

func Equals(v1, v2 interface{}) bool {
	list1, isList1 := v1.([]interface{})
	list2, isList2 := v2.([]interface{})
	if isList1 && isList2 {
		return sliceEquals(list1, list2)
	}

	vect1, isVect1 := v1.(Vector)
	vect2, isVect2 := v2.(Vector)
	if isVect1 && isVect2 {
		return sliceEquals(vect1, vect2)
	}

	map1, isMap1 := v1.(map[interface{}]interface{})
	map2, isMap2 := v2.(map[interface{}]interface{})
	if isMap1 && isMap2 {
		return mapEquals(map1, map2)
	}

	return v1 == v2
}

func sliceEquals(vect1, vect2 []interface{}) bool {
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

func mapEquals(map1, map2 map[interface{}]interface{}) bool {
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
