package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/openschoolcn/zfn-api-go/common"
)

type ClientOptions struct {
	BaseURL string // base url of the api
	// CourseSchedule          [][]string // course schedule
	// IgnoreCourses           []string   // courses to ignore
	// DetailedCategoryCourses []string   // courses to be access for detailed category
}

type Client struct {
	ClientOptions
	r       *resty.Client
	Cookies []*http.Cookie    // cookies to be used for the requests
	Headers map[string]string // headers to be used for the requests
	Timeout int               // timeout for the requests
}

func (c *Client) Get(url string, query map[string]string) (resp *resty.Response, err error) {
	c.r.SetTimeout(time.Duration(c.Timeout) * time.Second)
	c.r.SetHeaders(c.Headers)
	for _, cookie := range c.Cookies {
		c.r.SetCookie(cookie)
	}
	if query == nil {
		return c.r.R().Get(url)
	}
	return c.r.R().SetQueryParams(query).Get(url)
}

func (c *Client) Post(url string, data map[string]string, noRedirect bool) (resp *resty.Response, err error) {
	c.r.SetTimeout(time.Duration(c.Timeout) * time.Second)
	c.r.SetHeaders(c.Headers)
	if noRedirect {
		c.r.SetRedirectPolicy(resty.NoRedirectPolicy())
	} else {
		c.r.SetRedirectPolicy()
	}
	for _, cookie := range c.Cookies {
		c.r.SetCookie(cookie)
	}
	resp, err = c.r.R().SetFormData(data).Post(url)
	if err != nil && strings.Contains(err.Error(), "auto redirect is disabled") {
		err = nil
	}

	return resp, err
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
		r:             resty.New(),
		Cookies:       cookies,
		Headers:       headers,
		Timeout:       timeout,
	}
}
