package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/sysu-yunz/doubanAnalysis/config"
	"github.com/sysu-yunz/doubanAnalysis/db"
	"github.com/sysu-yunz/doubanAnalysis/global"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	err error
)

func init() {
	if err != nil {
		log.Fatal("Init Bot %+v", err)
	}

	mgoPwd := config.ViperEnvVariable("MGO_PWD")
	global.MgoDB = db.NewDB(mgoPwd)
}


func main() {
	ms := getMovies()
	global.MgoDB.InsertMoviesBasic(ms)
}

func getMovies() []db.Movie {

	var ms []db.Movie

	url := `https://movie.douban.com/people/`+config.ViperEnvVariable("DoubanID")+`/collect`
	doc := getHTML(url, `div[class="info"]`)
	start := findBasicSubjectInfo(doc)
	ms = append(ms, start...)

	totalPage := findTotalNum(doc)/15 + 1
	// totalPage := 1
	// 翻页-组装URL
	// https://movie.douban.com/people/dukeyunz/collect?start=15&sort=time&rating=all&filter=all&mode=grid
	// 因为第一页已经跑过一次了，直接从第二页开始

	for i := 1; i < totalPage; i++ {
		currentURL := fmt.Sprintf("%s?start=%d&sort=time&rating=all&filter=all&mode=grid", url, i*15)
		currentDoc := getHTML(currentURL, `div[class="info"]`)
		ms = append(ms, findBasicSubjectInfo(currentDoc)...)
	}

	return ms
}

func getHTML(url string, wait interface{}) *goquery.Document {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless",true),
		chromedp.Flag("blink-settings","imageEnable=false"),
		chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko)`),
	}

	c,_ := chromedp.NewExecAllocator(context.Background(),options...)

	chromeCtx, cancel := chromedp.NewContext(c,chromedp.WithLogf(log.Printf))
	_ = chromedp.Run(chromeCtx, make([]chromedp.Action, 0, 1)...)

	timeOutCtx, cancel := context.WithTimeout(chromeCtx,60*time.Second)
	defer cancel()

	var htmlContent string

	log.Printf("chrome visit page %s\n",url)
	err := chromedp.Run(timeOutCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(wait),
		chromedp.OuterHTML(`document.querySelector("body")`,&htmlContent,chromedp.ByJSPath),
	)
	if err!=nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func findSubjectRunTime(doc *goquery.Document) {

	//doc.Find("#info").Each(func(i int, s *goquery.Selection) {
	//	op, _ := s.Attr("property")
	//	con, _ := s.Attr("content")
	//	if op == "v:runtime" {
	//		fmt.Println(con)
	//	}
	//})

	var isMark bool
	doc.Find("div#info").Contents().Each(func(i int, s *goquery.Selection) {
		if s.Text() == "片长:" {
			fmt.Printf("片长:%s\n", s.Next().Text())
		}

		if s.Text() == "集数:" {
			goquery.NodeName(s.Next())
			fmt.Printf("集数: ")
			isMark = true
		}
		if s.Text() == "单集片长:" {
			fmt.Printf("单集片长:")
			isMark = true
		}

		if goquery.NodeName(s) == "#text" && isMark {
			fmt.Println(s.Text())
			isMark = false
		}
	})
}

func findTotalNum(doc *goquery.Document) int {
	s := doc.Find("h1").Text()
	re := regexp.MustCompile(`(?s)\((.*)\)`)
	m := re.FindAllStringSubmatch(s,-1)
	fmt.Printf(m[0][1])

	if num, err := strconv.Atoi(m[0][1]); err == nil {
		return num
	}

	return 0
}

func findBasicSubjectInfo(doc *goquery.Document) []db.Movie {
	// 获取内容
	// title 标题
	// <li class="title">
	//                        <a href="https://movie.douban.com/subject/26413293/" class="">
	//                            <em>大秦赋</em>
	//                             / 大秦帝国4：东出 / 大秦帝国之东出
	//                        </a>
	//                            <span class="playable">[可播放]</span>
	//                    </li>

	// link 链接
	// rate 评分
	// <span class="rating1-t"></span>

	// date 日期
	// <span class="date">2020-12-16</span>

	// comment 评论
	// <span class="comment">本来还说快进随便看看，弃剧了。以后国产剧一定放凉了再看，真nm坑。</span>

	// img 图片
	// <img alt="Warrior" src="https://img9.doubanio.com/view/photo/s_ratio_poster/public/p2619810129.webp" class="">

	var ms []db.Movie
	doc.Find(".item").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		img, _ := s.Find("img").Attr("src")
		titleSel := s.Find(".title")
		title := titleSel.Find("em").Text()
		link, _ := titleSel.Find("a").Attr("href")
		dateSel := s.Find(".date")
		rate, _ := dateSel.Prev().Attr("class")
		date := dateSel.Text()
		comment := s.Find(".comment").Text()
		la := strings.Split(link, "/")
		subject := la[len(la)-2]
		ms = append(ms, db.Movie{
			Subject: subject,
			Title: title,
			Link: link,
			Rate: rate,
			Date: date,
			Comment: comment,
			Img: img,
		})
	})

	return ms
}