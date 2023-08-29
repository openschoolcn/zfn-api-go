package common

import (
	"fmt"
	"log"
)

func CatchReqError(url string, result map[string]interface{}, err error) (map[string]interface{}, error) {
	log.Default().Printf(fmt.Sprintf("请求%s失败", url))
	result["code"] = 999
	result["msg"] = "请求失败"
	return result, err
}

func CatchLogicError(url string, msg string, err error) (map[string]interface{}, error) {
	log.Default().Printf(fmt.Sprintf("请求%s失败: %s", url, msg))
	result := make(map[string]interface{})
	result["code"] = 998
	result["msg"] = msg
	return result, err
}
