package common

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
