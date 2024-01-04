package insta

import (
	"bytes"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

type SnapInstaResponse struct {
	Username    string   `json:"username"`
	Avatar      string   `json:"avatar"`
	ResultMedia []string `json:"result_media"`
}

func NewCloudflareBypass() *resty.Client {
	httpTransport := &http.Client{}
	httpTransport.Transport = cloudflarebp.AddCloudFlareByPass(httpTransport.Transport)

	client := resty.New()
	client.SetTransport(httpTransport.Transport)

	return client
}

func GetSnapInsta(instagram string) (response *SnapInstaResponse, err error) {
	client := NewCloudflareBypass()
	resp, err := client.R().
		SetFormData(map[string]string{
			"url":    instagram,
			"action": "post",
			"lang":   "id",
		}).
		SetHeader("Origin", "https://snapinsta.app").
		SetHeader("Referer", "https://snapinsta.app/id").
		SetHeader("User-Agent", browser.Firefox()).
		Post("https://snapinsta.app/action2.php")
	if err != nil {
		return nil, err
	}

	defer resp.RawBody().Close()
	script := resp.String()
	splited := strings.Split(script, "}(")
	if len(splited) <= 1 {
		return nil, errors.New("[404] Could not find executable script")
	}
	splited = strings.Split(strings.Split(splited[1], "))")[0], ",")
	h := strings.ReplaceAll(splited[0], "\"", "")
	u, _ := strconv.Atoi(splited[1])
	n := strings.ReplaceAll(splited[2], "\"", "")
	t, _ := strconv.Atoi(splited[3])
	e, _ := strconv.Atoi(splited[4])
	r, _ := strconv.Atoi(splited[5])

	dec := DecodeSnap(h, u, n, t, e, r)
	html := innerHtml.FindStringSubmatch(dec)[1]
	parsedHtml := strings.ReplaceAll(html, `\"`, "")

	document, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(parsedHtml)))
	if err != nil {
		return response, err
	}

	response = &SnapInstaResponse{
		ResultMedia: make([]string, 0),
	}

	response.Username = document.Find("div.download-top > div").First().Text()
	avatar, _ := document.Find("div.download-top > div > img").First().Attr("src")
	response.Avatar = avatar

	document.Find("div.download-bottom > a").Each(func(i int, selection *goquery.Selection) {
		href, ok := selection.Attr("href")
		if ok {
			response.ResultMedia = append(response.ResultMedia, href)
		}
	})
	return response, nil
}

func DecodeSnap(h string, u int, n string, t int, e int, r int) string {
	decodedString := ""
	for i := 0; i < len(h); i++ {
		s := ""
		for h[i:i+1] != n[e:e+1] {
			s += h[i : i+1]
			i++
		}
		for j := 0; j < len(n); j++ {
			s = strings.ReplaceAll(s, n[j:j+1], strconv.Itoa(j))
		}
		chipResult, _ := strconv.Atoi(chip(s, e, 10))
		decodedString += string(rune(chipResult - t))
	}
	return decodedString
}

var innerHtml = regexp.MustCompile("\\.innerHTML = \"(.*?)\";")

// var innerToken = regexp.MustCompile("get_progressApi\\('/render\\.php\\?token=(.*?)'\\);")
func chip(d string, e int, f int) string {
	g := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"
	h := g[:e]
	i := g[:f]
	j := 0
	for c := len(d) - 1; c >= 0; c-- {
		b := string(d[c])
		index := strings.Index(h, b)
		if index != -1 {
			j += index * intPow(e, len(d)-1-c)
		}
	}
	k := ""
	for j > 0 {
		k = string(i[j%f]) + k
		j = (j - (j % f)) / f
	}
	if k == "" {
		return "0"
	}
	return k
}

func intPow(base, exponent int) int {
	result := 1
	for i := 0; i < exponent; i++ {
		result *= base
	}
	return result
}
