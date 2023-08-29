package common

import (
	"fmt"
	"log"

	"github.com/openschoolcn/zfn-api-go/models"
)

func CatchReqError(url string, err error) (models.Result, error) {
	log.Default().Printf(fmt.Sprintf("请求%s失败", url))
	result := models.Result{
		Code: 999,
		Msg:  "请求失败",
	}
	return result, err
}

func CatchLogicError(msg string, err error) (models.Result, error) {
	log.Default().Printf(fmt.Sprintf("错误: %s", msg))
	result := models.Result{
		Code: 998,
		Msg:  msg,
	}
	return result, err
}
