package api

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/openschoolcn/zfn-api-go/common"
	"github.com/openschoolcn/zfn-api-go/models"
)

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
	pubKey := common.Str2Map(keyResp.String())
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
		Data: models.LoginKaptcha{
			Sid:        sid,
			CsrfToken:  csrfToken,
			Cookies:    common.Cookie2Map(c.Cookies),
			Modulus:    modulus,
			Exponent:   exponent,
			KaptchaPic: kaptcha,
			Timestamp:  common.GetNowUnix(),
		}}, nil
}

func (c *Client) LoginWithKaptcha(loginKaptcha models.LoginKaptcha, password string, kaptcha string) (models.Result, error) {
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
		return common.CatchReqError(LoginURL, err)
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

func (c *Client) _Info() (models.Result, error) {
	infoResp, err := c.Get(common.UrlJoin(c.BaseURL, InfoURL2), map[string]string{"gnmkdm": "N100801"})
	if err != nil {
		return common.CatchReqError(InfoURL2, err)
	}
	bodyReader := bytes.NewReader(infoResp.Body())
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return common.CatchLogicError("解析个人信息页响应失败", err)
	}
	tips := doc.Find("h5").Text()
	if strings.Contains(tips, "用户登录") {
		return common.CatchCustomError(1006, "未登录或登录过期")
	}
	pendingResult := make(map[string]interface{})
	extractInfo := func(selection *goquery.Selection) (string, string) {
		content := selection.Find("div .form-group")
		key := strings.TrimSpace(content.Find("label").Eq(0).Text())
		value := strings.TrimSpace(content.Find("p").Text())
		if value == "" {
			value = strings.TrimSpace(content.Find("label").Eq(1).Text())
		}
		return key, value
	}
	// 学生基本信息
	doc.Find("div .col-sm-6").Each(func(i int, selection *goquery.Selection) {
		key, value := extractInfo(selection)
		pendingResult[key] = value
	})
	// 学生其它信息
	doc.Find("div .col-sm-4").Each(func(i int, selection *goquery.Selection) {
		key, value := extractInfo(selection)
		pendingResult[key] = value
	})
	if _, exists := pendingResult["学号："]; !exists {
		return models.Result{
			Code: 1014,
			Msg:  "当前学年学期无个人数据，您可能已经毕业了。如果是专升本同学，请使用专升本后的新学号登录～",
		}, nil
	}
	stuInfo := models.StudentInfo{
		Sid:             common.GetString(pendingResult, "学号："),
		Name:            common.GetString(pendingResult, "姓名："),
		Domicile:        common.GetString(pendingResult, "籍贯："),
		PhoneNumber:     common.GetString(pendingResult, "手机号码："),
		Email:           common.GetString(pendingResult, "电子邮箱："),
		PoliticalStatus: common.GetString(pendingResult, "政治面貌："),
		Nationality:     common.GetString(pendingResult, "民族："),
		CollegeName:     common.GetString(pendingResult, "学院名称："),
		MajorName:       common.GetString(pendingResult, "专业名称："),
		ClassName:       common.GetString(pendingResult, "班级名称："),
	}

	if stuInfo.CollegeName == "" {
		extraInfoReq, err := c.Post(common.UrlJoin(c.BaseURL, ExtraInfoURL), map[string]string{
			"offDetails": "1", "gnmkdm": "N106005", "czdmKey": "00",
		}, false)
		if err != nil {
			return common.CatchReqError(ExtraInfoURL, err)
		}
		bodyReader := bytes.NewReader(extraInfoReq.Body())
		doc, err := goquery.NewDocumentFromReader(bodyReader)
		if err != nil {
			return common.CatchLogicError("解析额外信息页响应失败", err)
		}
		if tips := doc.Find("p .error_title").Text(); tips != "无功能权限，" {
			// 通过学生证补办申请入口，来补全部分信息
			doc.Find("div .col-sm-6").Each(func(i int, selection *goquery.Selection) {
				key, value := extractInfo(selection)
				pendingResult[key] = value
			})
			stuInfo.CollegeName = common.GetString(pendingResult, "学院")
			stuInfo.MajorName = common.GetString(pendingResult, "专业")
			stuInfo.ClassName = common.GetString(pendingResult, "班级")
		}
	}
	return models.Result{
		Code: 1000,
		Msg:  "获取个人信息成功",
		Data: stuInfo,
	}, nil
}

func (c *Client) Info() (models.Result, error) {
	infoResp, err := c.Get(common.UrlJoin(c.BaseURL, InfoURL1), map[string]string{"gnmkdm": "N100801"})
	if err != nil {
		return common.CatchReqError(InfoURL1, err)
	}
	if infoResp.String() == "null" {
		return c._Info()
	}
	bodyReader := bytes.NewReader(infoResp.Body())
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return common.CatchLogicError("解析个人信息页响应失败", err)
	}
	tips := doc.Find("h5").Text()
	if strings.Contains(tips, "用户登录") {
		return common.CatchCustomError(1006, "未登录或登录过期")
	}
	info := infoResp.String()
	infoMap := common.Str2Map(info)
	stuInfo := models.StudentInfo{
		Sid:             common.GetString(infoMap, "xh"),
		Name:            common.GetString(infoMap, "xm"),
		Domicile:        common.GetString(infoMap, "jg"),
		PhoneNumber:     common.GetString(infoMap, "sjhm"),
		Email:           common.GetString(infoMap, "dzyx"),
		PoliticalStatus: common.GetString(infoMap, "zzmm"),
		Nationality:     common.GetString(infoMap, "mzm"),
		CollegeName:     common.GetString(infoMap, "zsjg_id", "jg_id"),
		MajorName:       common.GetString(infoMap, "zszyh_id", "zyh_id"),
		ClassName:       common.GetString(infoMap, "bh_id", "xjztdm"),
	}
	return models.Result{
		Code: 1000,
		Msg:  "获取个人信息成功",
		Data: stuInfo,
	}, nil
}

func (c *Client) Grade(year int, term int, usePersonalGrade bool) (models.Result, error) {
	gradeURL := GradeURL
	if usePersonalGrade {
		gradeURL = PersonalGradeURL
	}
	reqTerm := ""
	if term != 0 {
		reqTerm = strconv.Itoa(term * term * 3)
	}
	gradeResp, err := c.Post(common.UrlJoin(c.BaseURL, gradeURL), map[string]string{
		"xnm":                    strconv.Itoa(year),
		"xqm":                    reqTerm,
		"_search":                "false",
		"nd":                     common.GetNowUnix(),
		"queryModel.showCount":   "100",
		"queryModel.currentPage": "1",
		"queryModel.sortName":    "",
		"queryModel.sortOrder":   "asc",
		"time":                   "0",
	}, false)
	if err != nil {
		return common.CatchReqError(gradeURL, err)
	}
	bodyReader := bytes.NewReader(gradeResp.Body())
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return common.CatchLogicError("解析成绩页响应失败", err)
	}
	if tips := doc.Find("h5").Text(); tips == "用户登录" {
		return common.CatchCustomError(1006, "未登录或登录过期")
	}
	grade := gradeResp.String()
	gradeMap := common.Str2Map(grade)
	gradeInfo := models.GradeInfo{
		Year: year,
		Term: term,
	}
	gradeCourse := gradeMap["items"].([]interface{})
	for _, item := range gradeCourse {
		gradeCourseMap := item.(map[string]interface{})
		gradeInfo.Courses = append(gradeInfo.Courses, models.GradeCourse{
			CourseId:        common.GetString(gradeCourseMap, "kch_id"),
			Title:           common.GetString(gradeCourseMap, "kcmc"),
			Teacher:         common.GetString(gradeCourseMap, "jsxm"),
			ClassName:       common.GetString(gradeCourseMap, "jxbmc"),
			Credit:          common.GetString(gradeCourseMap, "xf"),
			Category:        common.GetString(gradeCourseMap, "kclbmc"),
			Nature:          common.GetString(gradeCourseMap, "kcxzmc"),
			Grade:           common.GetString(gradeCourseMap, "cj"),
			GradePoint:      common.GetString(gradeCourseMap, "jd"),
			GradeNature:     common.GetString(gradeCourseMap, "ksxz"),
			TeachingCollege: common.GetString(gradeCourseMap, "kkbmmc"),
			Mark:            common.GetString(gradeCourseMap, "kcbj"),
		})
	}
	gradeInfo.Count = len(gradeCourse)
	return models.Result{
		Code: 1000,
		Msg:  "获取成绩成功",
		Data: gradeInfo,
	}, nil
}
