package main

func First(array []interface{}) {
	return array[0]
}

func Last(array []interface{}) interface{} {
	length := len(array)
	if length == 0 {
		return nil
	}

	return array[length-1]
}
