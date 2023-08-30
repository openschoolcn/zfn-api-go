package common

import (
	"strconv"
)

func GetString(result map[string]interface{}, keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	for _, key := range keys {
		if val, exists := result[key].(string); exists {
			return val
		}
	}
	return ""
}

func GetInt(result map[string]interface{}, keys ...string) int {
	if len(keys) == 0 {
		return 0
	}
	for _, key := range keys {
		if val, exists := result[key].(string); exists {
			conv, _ := strconv.Atoi(val)
			return conv
		}
	}
	return 0
}

func GetFloat(result map[string]interface{}, keys ...string) float64 {
	if len(keys) == 0 {
		return 0
	}
	for _, key := range keys {
		if val, exists := result[key].(string); exists {
			conv, _ := strconv.ParseFloat(val, 64)
			return conv
		}
	}
	return 0
}
