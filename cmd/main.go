package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/openschoolcn/zfn-api-go/api"
	"github.com/openschoolcn/zfn-api-go/common"
	"github.com/openschoolcn/zfn-api-go/models"
	"github.com/spf13/cobra"
)

var client *api.Client = api.NewClient(api.ClientOptions{}, nil, 5)

type Config struct {
	BaseURL  string `json:"base_url"`
	Sid      string `json:"sid"`
	Password string `json:"password"`
}

func display(result any) {
	json, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	fmt.Println(string(json))
}

func initConf() Config {
	configFile := "config.json"
	if _, err := os.Stat(configFile); err != nil {
		initConfig := Config{
			BaseURL:  "",
			Sid:      "",
			Password: "",
		}
		configBytes, _ := json.Marshal(initConfig)
		os.WriteFile(configFile, configBytes, 0644)
		return initConfig
	}
	configReader, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("读取配置文件错误:", err)
	}
	conf := common.Str2Map(string(configReader))
	return Config{
		BaseURL:  conf["base_url"].(string),
		Sid:      conf["sid"].(string),
		Password: conf["password"].(string),
	}
}

func setConf(config Config) {
	configBytes, _ := json.Marshal(config)
	os.WriteFile("config.json", configBytes, 0644)
}

func setCookies(cookies []*http.Cookie) {
	cookiesStr, _ := json.Marshal(common.Cookie2Map(cookies))
	cookiesFile := "cookies.json"
	os.WriteFile(cookiesFile, []byte(cookiesStr), 0644)
}

func getCookies() []*http.Cookie {
	cookiesFile := "cookies.json"
	if _, err := os.Stat(cookiesFile); err != nil {
		cookies := []*http.Cookie{}
		cookiesMap := common.Map2Cookie(map[string]string{})
		cookiesBytes, _ := json.Marshal(cookiesMap)
		os.WriteFile(cookiesFile, cookiesBytes, 0644)
		return cookies
	}
	cookiesStr, _ := os.ReadFile("cookies.json")
	cookies := make(map[string]string)
	json.Unmarshal([]byte(cookiesStr), &cookies)
	return common.Map2Cookie(cookies)
}

func main() {
	var config Config = initConf()
	cookies := getCookies()
	client = api.NewClient(api.ClientOptions{BaseURL: config.BaseURL}, cookies, 5)
	var rootCmd = &cobra.Command{Use: "zfn-cli"}

	var cmdConfig = &cobra.Command{
		Use:   "config",
		Short: "Configure",
		Run: func(cmd *cobra.Command, args []string) {
			baseURL, _ := cmd.Flags().GetString("base_url")
			config.BaseURL = baseURL
			setConf(config)
			fmt.Printf("Base URL is set to: %s\n", baseURL)
		},
	}

	interfaceLogin := func(sid string, password string) {
		client = api.NewClient(api.ClientOptions{BaseURL: config.BaseURL}, nil, 5)
		result, err := client.Login(sid, password)
		if err != nil {
			log.Fatal(err)
		}
		if result.Code == 1001 {
			loginKaptcha := result.Data.(models.LoginKaptcha)
			DisplayCaptcha(loginKaptcha.KaptchaPic)
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
		}
		config.Sid = sid
		config.Password = password
		setConf(config)
		if result.Data != nil {
			cookiesMap := result.Data.(map[string]interface{})["cookies"].(map[string]string)
			cookies = common.Map2Cookie(cookiesMap)
			setCookies(cookies)
		}
		display(result)
	}

	var cmdLogin = &cobra.Command{
		Use:   "login",
		Short: "Login with user and password",
		Run: func(cmd *cobra.Command, args []string) {
			sid, _ := cmd.Flags().GetString("sid")
			password, _ := cmd.Flags().GetString("password")
			fmt.Printf("Logging in with sid: %s and password: %s\n", sid, password)
			interfaceLogin(sid, password)
		},
	}

	var cmdInfo = &cobra.Command{
		Use:   "info",
		Short: "Get student info",
		Run: func(cmd *cobra.Command, args []string) {
			result, err := client.Info()
			if err != nil {
				log.Fatal(err)
			}
			if result.Code == 1006 {
				interfaceLogin(config.Sid, config.Password)
				result, err = client.Info()
				if err != nil {
					log.Fatal(err)
				}
			}
			display(result)
		},
	}

	var cmdGrade = &cobra.Command{
		Use:   "grade",
		Short: "Get grade by year and term",
		Run: func(cmd *cobra.Command, args []string) {
			year, _ := cmd.Flags().GetInt("year")
			term, _ := cmd.Flags().GetInt("term")
			fmt.Printf("Fetching grade for year: %d and term: %d\n", year, term)
			result, err := client.Grade(year, term, true)
			if err != nil {
				log.Fatal(err)
			}
			if result.Code == 1006 {
				interfaceLogin(config.Sid, config.Password)
				result, err = client.Grade(year, term, true)
				if err != nil {
					log.Fatal(err)
				}
			}
			display(result)
		},
	}

	cmdConfig.Flags().StringP("base_url", "b", "", "Base URL for API")
	cmdLogin.Flags().StringP("sid", "u", "", "sid for login")
	cmdLogin.Flags().StringP("password", "p", "", "Password for login")
	cmdGrade.Flags().Int("year", 2022, "Year for fetching grade")
	cmdGrade.Flags().Int("term", 0, "Term for fetching grade")

	rootCmd.AddCommand(cmdConfig, cmdLogin, cmdInfo, cmdGrade)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
