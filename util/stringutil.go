package util

// StringInSlice 判断字符串是否在slice中
func StringInSlice(items []string, item string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}
