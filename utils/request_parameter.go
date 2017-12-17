package utils

import "strconv"

// GetIntOrDefault return value as string or if value is empty or not string return defaultValue
func GetIntOrDefault(value string, defaultValue int) int {
	if len(value) == 0 {
		return defaultValue
	}
	ret, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return ret
}
