package common

import (
	"strconv"
	"time"
)

func GetNowUnix() string {
	now := time.Now().Unix()
	return strconv.FormatInt(now, 13)
}
