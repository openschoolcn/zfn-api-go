package main

import (
	"testing"

	"github.com/openschoolcn/zfn-api-go/common"
)

func TestUrlJoin(t *testing.T) {
	baseURL := "https://jwxt.xcc.edu.cn/jwxtag/"
	expected := "https://jwxt.xcc.edu.cn/jwxtag/xtgl/login_slogin.html"
	result := common.UrlJoin(baseURL, "xtgl/login_slogin.html")
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestBody2Map(t *testing.T) {
	jsonBody := `{
		"success": true,
		"modulus": "asdasdasdasdgadfa==",
		"content": "你好我好大家好"
	}
	`
	expected := map[string]interface{}{
		"success": true,
		"modulus": "asdasdasdasdgadfa==",
		"content": "你好我好大家好",
	}
	result := common.Str2Map(jsonBody)
	if result["success"] != expected["success"] {
		t.Errorf("Expected %v, got %v", expected["success"], result["success"])
	}
}
