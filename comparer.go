package main

func Equals(v1, v2 any) bool {
	list1, isList1 := v1.(List)
	list2, isList2 := v2.(List)
	if isList1 && isList2 {
		return sliceEquals(list1, list2)
	}

	vect1, isVect1 := v1.([]any)
	vect2, isVect2 := v2.([]any)
	if isVect1 && isVect2 {
		return sliceEquals(vect1, vect2)
	}

	map1, isMap1 := v1.(map[any]any)
	map2, isMap2 := v2.(map[any]any)
	if isMap1 && isMap2 {
		return mapEquals(map1, map2)
	}

	return v1 == v2
}

func sliceEquals(slice1, slice2 []any) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := 0; i < len(slice1); i++ {
		if !Equals(slice1[i], slice2[i]) {
			return false
		}
	}
	return true
}

func mapEquals(map1, map2 map[any]any) bool {
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
