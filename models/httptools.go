package models

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

var (
	client *http.Client
)

func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}
	client = &http.Client{
		Transport: &http.Transport{
			/* debug tool
			Proxy: func(request *http.Request) (url *url.URL, e error) {
				return url.Parse("http://127.0.0.1:8888")
			},
			*/

		},
		Jar: jar,
	}
}

func GetStrings(url string) string {
	bytes, err := GetBytes(url)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func PostJson(url string, obj string) string {
	req, err := http.NewRequest("POST", url, strings.NewReader(obj))
	if err != nil {
		return ""
	}
	req.Header["User-Agent"] = []string{"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:67.0) Gecko/20100101 Firefox/67.0"}
	req.Header["Content-type"] = []string{"application/json;odata=verbose"}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return ""
	}
	return string(body)
}

func GetBytes(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header["User-Agent"] = []string{"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:67.0) Gecko/20100101 Firefox/67.0"}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
