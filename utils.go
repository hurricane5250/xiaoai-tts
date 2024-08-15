package goxiaoai

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

func ParseToekn(source string) string {
	c := regexp.MustCompile("serviceToken=(.*?);")
	s := c.FindStringSubmatch(source)
	return s[len(s)-1]
}

func NewRequest(method string, url string, body io.Reader) (request *http.Request, err error) {
	request, err = http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	contentType := APPLICATION_JSON
	if request.Method == http.MethodPost {
		contentType = "application/x-www-form-urlencoded"
	}

	userAgent := APP_UA
	if strings.Contains(request.URL.Host, "mina.mi.com") {
		userAgent = MINA_UA
	}

	request.Header.Add("Content-Type", contentType)
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("User-Agent", userAgent)
	request.Header.Add("Accept", "*/*")

	return
}

func Sha1Base64(data string) string {
	o := sha1.New()
	o.Write([]byte(data))
	return fmt.Sprintf("%s=", base64.RawStdEncoding.EncodeToString(o.Sum(nil)))
}

func GetRandomString(n int) string {
	str := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	for i := 0; i < n; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}
