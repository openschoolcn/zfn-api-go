package api

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/openschoolcn/zfn-api-go/common"
)

const (
	KeyURL     = "xtgl/login_getPublicKey.html"
	LoginURL   = "xtgl/login_slogin.html"
	KaptchaURL = "kaptcha"
)

type ClientOptions struct {
	BaseURL string // base url of the api
	// CourseSchedule          [][]string // course schedule
	// IgnoreCourses           []string   // courses to ignore
	// DetailedCategoryCourses []string   // courses to be access for detailed category
}

type Client struct {
	ClientOptions
	Cookies []*http.Cookie    // cookies to be used for the requests
	Headers map[string]string // headers to be used for the requests
	Timeout int               // timeout for the requests
}

func (c *Client) Get(url string, query map[string]string) (resp *resty.Response, err error) {
	client := resty.New()
	client.SetTimeout(time.Duration(c.Timeout) * time.Second)
	client.SetHeaders(c.Headers)
	for _, cookie := range c.Cookies {
		client.SetCookie(cookie)
	}
	if query == nil {
		return client.R().Get(url)
	}
	return client.R().SetQueryParams(query).Get(url)
}

func (c *Client) Post(url string, data map[string]string) (resp *resty.Response, err error) {
	client := resty.New()
	client.SetTimeout(time.Duration(c.Timeout) * time.Second)
	client.SetHeaders(c.Headers)
	for _, cookie := range c.Cookies {
		client.SetCookie(cookie)
	}
	return client.R().SetFormData(data).Post(url)
}

func NewClient(options ClientOptions, cookies []*http.Cookie, timeout int) *Client {
	headers := map[string]string{
		"Referer":    common.UrlJoin(options.BaseURL, LoginURL),
		"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36",
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
	}
	if cookies == nil {
		cookies = []*http.Cookie{}
	}
	return &Client{
		ClientOptions: options,
		Cookies:       cookies,
		Headers:       headers,
		Timeout:       timeout,
	}
}

type LoginKaptcha struct {
	Sid        string            `json:"sid"`
	CsrfToken  string            `json:"csrf_token"`
	Cookies    map[string]string `json:"cookies"`
	Modulus    string            `json:"modulus"`
	Exponent   string            `json:"exponent"`
	KaptchaPic string            `json:"kaptcha_pic"`
	Timestamp  int64             `json:"timestamp"`
}

func (c *Client) Login(sid string, password string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	// get csrf_token

	csrfResp, err := c.Get(common.UrlJoin(c.BaseURL, LoginURL), nil)
	if err != nil {
		return common.CatchReqError(LoginURL, result, err)
	}
	bodyReader := bytes.NewReader(csrfResp.Body())
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return common.CatchLogicError(LoginURL, "解析csrf页响应失败", err)
	}
	csrfTokenSlection := doc.Find("input#csrftoken")
	if csrfTokenSlection.Length() == 0 {
		return common.CatchLogicError(LoginURL, "获取csrftoken失败", nil)
	}
	csrfToken, exists := csrfTokenSlection.Attr("value")
	if !exists {
		return common.CatchLogicError(LoginURL, "获取csrftoken失败", nil)
	}
	c.Cookies = csrfResp.Cookies()
	// get public key
	keyResp, err := c.Get(common.UrlJoin(c.BaseURL, KeyURL), nil)
	if err != nil {
		return common.CatchReqError(KeyURL, result, err)
	}
	pubKey := common.Body2Map(keyResp.String())
	modulus := pubKey["modulus"].(string)
	exponent := pubKey["exponent"].(string)
	yzm := doc.Find("#yzmDiv").Text()
	if yzm == "" {
		// no captcha
		encryptPassword, _ := common.EncryptPassword(password, modulus, exponent)
		loginMap := map[string]string{
			"csrftoken": csrfToken,
			"yhm":       sid,
			"mm":        encryptPassword,
		}
		loginResp, err := c.Post(common.UrlJoin(c.BaseURL, LoginURL), loginMap)
		if err != nil {
			return common.CatchReqError(LoginURL, result, err)
		}
		bodyReader := bytes.NewReader(loginResp.Body())
		doc, _ := goquery.NewDocumentFromReader(bodyReader)
		tips := doc.Find("p#tips").Text()
		if tips != "" {
			if strings.Contains(tips, "用户名或密码") {
				result["code"] = 1002
				result["msg"] = "用户名或密码错误"
				return result, nil
			}
			tips = strings.TrimSpace(tips)
			return common.CatchLogicError(LoginURL, tips, nil)
		}
		c.Cookies = loginResp.Cookies()
		result["code"] = 1000
		result["msg"] = "登录成功"
		result["data"] = map[string]interface{}{
			"cookies": common.Cookie2Map(c.Cookies),
		}
		return result, nil
	}
	// require captcha
	kaptchaResp, err := c.Get(common.UrlJoin(c.BaseURL, KaptchaURL), nil)
	if err != nil {
		return common.CatchReqError(KaptchaURL, result, err)
	}
	kaptcha := common.Base64Encode(kaptchaResp.Body())
	result["code"] = 1001
	result["msg"] = "获取验证码成功"
	result["data"] = LoginKaptcha{
		Sid:        sid,
		CsrfToken:  csrfToken,
		Cookies:    common.Cookie2Map(c.Cookies),
		Modulus:    modulus,
		Exponent:   exponent,
		KaptchaPic: kaptcha,
		Timestamp:  time.Now().Unix(),
	}
	return result, nil
}

func (c *Client) LoginWithKaptcha(loginKaptcha LoginKaptcha, password string, kaptcha string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	encryptPassword, _ := common.EncryptPassword(password, loginKaptcha.Modulus, loginKaptcha.Exponent)
	loginMap := map[string]string{
		"csrftoken": loginKaptcha.CsrfToken,
		"yhm":       loginKaptcha.Sid,
		"mm":        encryptPassword,
		"yzm":       kaptcha,
	}
	c.Cookies = common.Map2Cookie(loginKaptcha.Cookies)
	loginResp, err := c.Post(common.UrlJoin(c.BaseURL, LoginURL), loginMap)
	if err != nil {
		return common.CatchReqError(LoginURL, result, err)
	}
	bodyReader := bytes.NewReader(loginResp.Body())
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return common.CatchLogicError(LoginURL, "解析登录页响应失败", err)
	}
	tips := doc.Find("p#tips").Text()
	if tips != "" {
		if strings.Contains(tips, "用户名或密码") {
			result["code"] = 1002
			result["msg"] = "用户名或密码错误"
			return result, nil
		}
		if strings.Contains(tips, "验证码") {
			result["code"] = 1004
			result["msg"] = "验证码错误"
			return result, nil
		}
		return common.CatchLogicError(LoginURL, tips, nil)
	}
	c.Cookies = loginResp.Cookies()
	cookies := common.Cookie2Map(c.Cookies)
	// 不同学校SSO不同可能导致需要的Cookie字段不同
	if _, exists := cookies["route"]; !exists {
		cookies["route"] = loginKaptcha.Cookies["route"]
		c.Cookies = common.Map2Cookie(cookies)
	}
	result["code"] = 1000
	result["msg"] = "登录成功"
	result["data"] = map[string]interface{}{
		"cookies": cookies,
	}
	return result, nil
}
