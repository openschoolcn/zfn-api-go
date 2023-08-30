package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/openschoolcn/zfn-api-go/api"
	"github.com/openschoolcn/zfn-api-go/common"
)

const (
	BaseURL  = "https://xxx.cn/"
	Sid      = "your sid"
	Password = "your password"
)

func main() {
	cookies := []*http.Cookie{}
	// cookies := []*http.Cookie{
	// 	{
	// 		Name:  "JSESSIONID",
	// 		Value: "922064B642351A73E4XC8A53B99C115D",
	// 	},
	// 	{
	// 		Name:  "route",
	// 		Value: "9000cc9b13577537120983e690e03421",
	// 	},
	// }
	client := api.NewClient(api.ClientOptions{BaseURL: BaseURL}, cookies, 5)

	if len(cookies) == 0 {
		// login
		result, err := client.Login(Sid, Password)
		if err != nil {
			log.Fatal(err)
		}
		if result.Code == 1001 {
			loginKaptcha := result.Data.(api.LoginKaptcha)
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
			result, err = client.LoginWithKaptcha(loginKaptcha, Password, kaptcha)
			if err != nil {
				log.Fatal(err)
			}
			if result.Code != 1000 {
				log.Fatal(result)
			}
			fmt.Println(result)
		} else {
			fmt.Println(result)
		}
	}

	// get student info
	result, err := client.Info()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
