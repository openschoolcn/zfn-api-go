package main

import (
	"fmt"
	"log"
	"os"

	"github.com/openschoolcn/zfn-api-go/api"
	"github.com/openschoolcn/zfn-api-go/common"
)

func main() {
	client := api.NewClient(api.ClientOptions{BaseURL: "https://xxx.com/"}, nil, 5)
	sid := "your sid"
	password := "your password"
	result, err := client.Login(sid, password)
	if err != nil {
		log.Fatal(err)
	}
	if result["code"].(int) == 1001 {
		loginKaptcha := result["data"].(api.LoginKaptcha)
		data, err := common.Base64Decode(loginKaptcha.KaptchaPic)
		if err != nil {
			log.Fatal(err)
		}
		os.WriteFile("kaptcha.png", data, 0644)
		fmt.Println("请输入验证码：")
		var kaptcha string
		_, err = fmt.Scanln(&kaptcha)
		if err != nil {
			log.Fatal(err)
		}
		result, err = client.LoginWithKaptcha(loginKaptcha, password, kaptcha)
		if err != nil {
			log.Fatal(err)
		}
		if result["code"].(int) != 1000 {
			log.Fatal(result)
		}
		fmt.Println(result)
	} else {
		fmt.Println(result)
	}
}
