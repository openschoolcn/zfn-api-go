package api

import (
	"bytes"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/openschoolcn/zfn-api-go/common"
	"github.com/openschoolcn/zfn-api-go/models"
)


type LoginKaptcha struct {
	Sid        string            `json:"sid"`
	CsrfToken  string            `json:"csrf_token"`
	Cookies    map[string]string `json:"cookies"`
	Modulus    string            `json:"modulus"`
	Exponent   string            `json:"exponent"`
	KaptchaPic string            `json:"kaptcha_pic"`
	Timestamp  int64             `json:"timestamp"`
}

func (c *Client) Login(sid string, password string) (models.Result, error) {
	// get csrf_token

	csrfResp, err := c.Get(common.UrlJoin(c.BaseURL, LoginURL), nil)
	if err != nil {
		return common.CatchReqError(LoginURL, err)
	}
	bodyReader := bytes.NewReader(csrfResp.Body())
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return common.CatchLogicError("解析csrf页响应失败", err)
	}
	csrfTokenSlection := doc.Find("input#csrftoken")
	if csrfTokenSlection.Length() == 0 {
		return common.CatchLogicError("获取csrftoken失败", nil)
	}
	csrfToken, exists := csrfTokenSlection.Attr("value")
	if !exists {
		return common.CatchLogicError("获取csrftoken失败", nil)
	}
	c.Cookies = csrfResp.Cookies()
	// get public key
	keyResp, err := c.Get(common.UrlJoin(c.BaseURL, KeyURL), nil)
	if err != nil {
		return common.CatchReqError(KeyURL, err)
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
		loginResp, err := c.Post(common.UrlJoin(c.BaseURL, LoginURL), loginMap, true)
		if err != nil {
			return common.CatchReqError(LoginURL, err)
		}
		bodyReader := bytes.NewReader(loginResp.Body())
		doc, _ := goquery.NewDocumentFromReader(bodyReader)
		tips := doc.Find("p#tips").Text()
		if tips != "" {
			if strings.Contains(tips, "用户名或密码") {
				return models.Result{
					Code: 1002,
					Msg:  "用户名或密码错误",
				}, nil
			}
			tips = strings.TrimSpace(tips)
			return common.CatchLogicError(tips, nil)
		}
		c.Cookies = loginResp.Cookies()
		return models.Result{
			Code: 1000,
			Msg:  "登录成功",
			Data: map[string]interface{}{
				"cookies": common.Cookie2Map(c.Cookies),
			},
		}, nil
	}
	// require captcha
	kaptchaResp, err := c.Get(common.UrlJoin(c.BaseURL, KaptchaURL), nil)
	if err != nil {
		return common.CatchReqError(KaptchaURL, err)
	}
	kaptcha := common.Base64Encode(kaptchaResp.Body())
	return models.Result{
		Code: 1001,
		Msg:  "获取验证码成功",
		Data: LoginKaptcha{
			Sid:        sid,
			CsrfToken:  csrfToken,
			Cookies:    common.Cookie2Map(c.Cookies),
			Modulus:    modulus,
			Exponent:   exponent,
			KaptchaPic: kaptcha,
			Timestamp:  time.Now().Unix(),
		}},nil
}

func (c *Client) LoginWithKaptcha(loginKaptcha LoginKaptcha, password string, kaptcha string) (models.Result, error) {
	encryptPassword, _ := common.EncryptPassword(password, loginKaptcha.Modulus, loginKaptcha.Exponent)
	loginMap := map[string]string{
		"csrftoken": loginKaptcha.CsrfToken,
		"yhm":       loginKaptcha.Sid,
		"mm":        encryptPassword,
		"yzm":       kaptcha,
	}
	c.Cookies = common.Map2Cookie(loginKaptcha.Cookies)
	loginResp, err := c.Post(common.UrlJoin(c.BaseURL, LoginURL), loginMap, true)
	if err != nil {
		return common.CatchReqError(LoginURL,  err)
	}
	bodyReader := bytes.NewReader(loginResp.Body())
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return common.CatchLogicError("解析登录页响应失败", err)
	}
	tips := doc.Find("p#tips").Text()
	if tips != "" {
		if strings.Contains(tips, "用户名或密码") {
			return models.Result{
				Code: 1002,
				Msg:  "用户名或密码错误",
			}, nil
		}
		if strings.Contains(tips, "验证码") {
			return models.Result{
				Code: 1004,
				Msg:  "验证码错误",
			}, nil
		}
		return common.CatchLogicError(tips, nil)
	}
	c.Cookies = loginResp.Cookies()
	cookies := common.Cookie2Map(c.Cookies)
	// 不同学校SSO不同可能导致需要的Cookie字段不同
	if _, exists := cookies["route"]; !exists {
		cookies["route"] = loginKaptcha.Cookies["route"]
		c.Cookies = common.Map2Cookie(cookies)
	}
	return models.Result{
		Code: 1000,
		Msg:  "登录成功",
		Data: map[string]interface{}{
			"cookies": cookies,
		},
	}, nil
}
