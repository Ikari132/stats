package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/robfig/cron/v3"
)

type Users struct {
	Twity int
	Wtc   int
}

func main() {
	c := cron.New()

	u := Users{
		Twity: 0,
		Wtc:   0,
	}

	u.Wtc = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/working-time-counter/lakpjellnlajgbedjhejhdbkmphhfolo?utm_source=chrome-ntp-icon&hl=en")
	u.Twity = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/twity/mnencakcmnofdmpgmklkknklikgpodoo?utm_source=chrome-ntp-icon&hl=en")
	sum := u.Wtc + u.Twity

	c.AddFunc("@every 3h", func() {
		u.Wtc = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/working-time-counter/lakpjellnlajgbedjhejhdbkmphhfolo?utm_source=chrome-ntp-icon&hl=en")
		u.Twity = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/twity/mnencakcmnofdmpgmklkknklikgpodoo?utm_source=chrome-ntp-icon&hl=en")

		sum = u.Wtc + u.Twity
		fmt.Println("Users count from cron", sum)
	})
	c.Start()

	r := gin.Default()
	r.Use(CORSMiddleware())

	r.GET("/stats", func(con *gin.Context) {
		con.JSON(http.StatusOK, gin.H{
			"users": sum, "usersByProduct": u,
		})
	})

	r.Run(":80")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getChromeExtensionUsers(extUrl string) chan int {
	r := make(chan int)

	go func() {
		c := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
			colly.AllowedDomains("chrome.google.com", "www.chrome.google.com"),
		)
		cookie := &http.Cookie{
			Name:   "CONSENT",
			Value:  "YES+cb.20211111-08-p0.en+FX+342",
			MaxAge: 300,
		}
		urlObj, _ := url.Parse("https://chrome.google.com/")
		j, err := cookiejar.New(nil)
		if err == nil {
			j.SetCookies(urlObj, []*http.Cookie{cookie})
			c.SetCookieJar(j)
		}

		c.OnRequest(func(r *colly.Request) {
			fmt.Println("Visiting", r.URL)
		})

		c.OnError(func(r *colly.Response, err error) {
			fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		})

		c.OnHTML("body", func(e *colly.HTMLElement) {
			doc, _ := goquery.NewDocumentFromReader(
				strings.NewReader(e.Text),
			)
			span := doc.Find("span[title]")
			users, _ := strconv.Atoi(strings.Split(span.Text(), " ")[0])

			fmt.Println("Users count", users)
			r <- users

		})
		c.Visit(extUrl)
	}()

	return r
}
