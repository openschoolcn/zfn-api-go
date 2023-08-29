package common

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

func Body2Map(body string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return result
}

func Cookie2Map(cookies []*http.Cookie) map[string]string {
	result := make(map[string]string)
	for _, cookie := range cookies {
		result[cookie.Name] = cookie.Value
	}
	return result
}

func Map2Cookie(cookies map[string]string) []*http.Cookie {
	result := make([]*http.Cookie, 0)
	for k, v := range cookies {
		result = append(result, &http.Cookie{Name: k, Value: v})
	}
	return result
}

func Base64Encode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func Base64Decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}
