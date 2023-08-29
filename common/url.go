package common

import "net/url"

func UrlJoin(baseURL string, path string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	newURL, err := url.Parse(path)
	if err != nil {
		return ""
	}
	return u.ResolveReference(newURL).String()
}
