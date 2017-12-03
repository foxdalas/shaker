package shaker

import "net/http"

func Fetch(url string) (resp *http.Response, err error) {
	return http.Get(url)
}
