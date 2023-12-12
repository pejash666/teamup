package util

func BoolToDB(anyBool bool) int {
	if anyBool {
		return 1
	} else {
		return 0
	}
}
