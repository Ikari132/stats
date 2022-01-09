package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Users struct {
	Twity int
	Wtc   int
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	c := cron.New()

	dbHost := os.Getenv("POSTGRES_HOST")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, dbUser, dbName, dbPassword)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db.AutoMigrate(&Log{}, &Product{})

	if err != nil {
		panic(err)
	}

	u := Users{
		Twity: 0,
		Wtc:   0,
	}

	u.Wtc = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/working-time-counter/lakpjellnlajgbedjhejhdbkmphhfolo?utm_source=chrome-ntp-icon&hl=en")
	u.Twity = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/twity/mnencakcmnofdmpgmklkknklikgpodoo?utm_source=chrome-ntp-icon&hl=en")

	sum := u.Wtc + u.Twity

	updateProductsCount(*db, u)
	addProductsLogs(*db, u)

	c.AddFunc("@every 3h", func() {
		u.Wtc = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/working-time-counter/lakpjellnlajgbedjhejhdbkmphhfolo?utm_source=chrome-ntp-icon&hl=en")
		u.Twity = <-getChromeExtensionUsers("https://chrome.google.com/webstore/detail/twity/mnencakcmnofdmpgmklkknklikgpodoo?utm_source=chrome-ntp-icon&hl=en")

		sum = u.Wtc + u.Twity

		updateProductsCount(*db, u)
		addProductsLogs(*db, u)

		fmt.Println("Users count from cron", sum)
	})
	c.Start()

	r := gin.Default()
	r.Use(CORSMiddleware())

	r.GET("/stats", func(con *gin.Context) {
		var logs []Log
		if err := db.Order("created_at desc").Find(&logs).Error; err != nil {
			panic(err)
		}

		con.JSON(http.StatusOK, gin.H{
			"users": sum, "usersByProduct": u, "history": logs,
		})
	})
	r.GET("/health", func(con *gin.Context) {
		con.JSON(http.StatusOK, gin.H{
			"status": "ok",
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
func updateProductsCount(db gorm.DB, u Users) {
	updateOrCreate(db, "Wtc", u.Wtc)
	updateOrCreate(db, "Twity", u.Twity)
}
func updateOrCreate(db gorm.DB, Name string, Count int) {
	var t Product
	db.Where(Product{Name: Name}).FirstOrCreate(&t)
	t.Count = Count
	db.Save(&t)
}
func addProductsLogs(db gorm.DB, u Users) {
	wtcLog := Log{
		Count:   u.Wtc,
		Product: "Wtc",
	}
	twityLog := Log{
		Count:   u.Twity,
		Product: "Twity",
	}
	db.Create(&wtcLog)
	db.Create(&twityLog)
}
