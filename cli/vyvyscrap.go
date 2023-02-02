package main

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type MangaInfo struct {
	title     string
	views     uint
	genres    []string
	rating    string
	coverUrl  string
	completed bool
	length    uint
	url       string
}

var LYP, LSAIP int

func ScrapVyvy() (*MangaInfo, error) {
	if LYP == 0 {
		lastYaoiPage, err := getGenreLastPage("yaoi")
		if err != nil {
			return nil, err
		}
		LYP = lastYaoiPage
	}
	if LSAIP == 0 {
		lastSAIPage, err := getGenreLastPage("shounen-ai")
		if err != nil {
			return nil, err
		}
		LSAIP = lastSAIPage
	}

	yaoiListCollector := colly.NewCollector(colly.AllowedDomains("vyvymanga.net"))

	BLList := []string{}
	yaoiListCollector.OnHTML("body > div.body > div > div.row.book-list > div .comic-item a", func(e *colly.HTMLElement) {
		if !IsEmptyString(e.Attr("href")) && IsEmptyString(e.ChildAttr("span.no-chapter", "class")) {
			BLList = append(BLList, e.Attr("href"))
		}
	})

	yaoiPage, err := rand.Int(rand.Reader, big.NewInt(int64(LYP+1)))
	if err != nil {
		return nil, err
	}
	yaoiListCollector.Visit("https://vyvymanga.net/genre/yaoi?page=" + yaoiPage.String())

	SAIPage, err := rand.Int(rand.Reader, big.NewInt(int64(LSAIP+1)))
	if err != nil {
		return nil, err
	}
	yaoiListCollector.Visit("https://vyvymanga.net/genre/shounen-ai?page=" + SAIPage.String())

	if len(BLList) == 0 {
		return nil, errors.New("no bls")
	}

	randYaoiIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(BLList))))
	if err != nil {
		return nil, err
	}

	randYaoiUrl := BLList[BigIntToInt(randYaoiIdx)]
	yaoiInfoCollector := colly.NewCollector(colly.AllowedDomains("vyvymanga.net"))

	mangaInfo := MangaInfo{genres: []string{}, url: "https://vyvymanga.net" + randYaoiUrl}

	yaoiInfoCollector.OnHTML("div.div-manga > div.row > div.col-md-7 > p:nth-child(7)", func(e *colly.HTMLElement) {
		formattedText := strings.ReplaceAll(strings.Replace(e.Text, "View:", "", 1), ",", "")
		if views, err := strconv.Atoi(formattedText); err == nil {
			mangaInfo.views = uint(views)
		}
	})
	yaoiInfoCollector.OnHTML("div.div-manga > div.row > div.col-md-7 > h1", func(e *colly.HTMLElement) {
		mangaInfo.title = e.Text
	})
	yaoiInfoCollector.OnHTML("div.div-manga > div.row > div.col-md-7 > p:nth-child(8) > a", func(e *colly.HTMLElement) {
		mangaInfo.genres = append(mangaInfo.genres, e.Text)
	})
	yaoiInfoCollector.OnHTML("div.div-manga > div.row > div.col-md-7 > p:nth-child(9)", func(e *colly.HTMLElement) {
		mangaInfo.rating = strings.Replace(e.Text, "Rating: ", "", 1)
	})
	yaoiInfoCollector.OnHTML("div.div-manga > div.row > div.col-md-5 > div > img", func(e *colly.HTMLElement) {
		mangaInfo.coverUrl = e.Attr("src")
	})
	yaoiInfoCollector.OnHTML("div.div-manga > div.row > div.col-md-7 > p:nth-child(6)", func(e *colly.HTMLElement) {
		mangaInfo.completed = !strings.Contains(strings.ToLower(e.Text), "ongoing")
	})
	yaoiInfoCollector.OnHTML("div.col-lg-8 > div.div-chapter.mt-5 > div.list > div > a", func(e *colly.HTMLElement) {
		mangaInfo.length++
	})

	yaoiInfoCollector.Visit("https://vyvymanga.net" + randYaoiUrl)

	return &mangaInfo, nil
}

func getGenreLastPage(genre string) (int, error) {
	c := colly.NewCollector(colly.AllowedDomains("vyvymanga.net"), colly.Async(true))

	respChan := make(chan int)
	errChan := make(chan error)

	c.OnHTML("body > div.body > div > div.d-flex.justify-content-center > nav > ul > li:nth-child(10) > a", func(e *colly.HTMLElement) {
		if IsEmptyString(e.Text) {
			errChan <- errors.New("Value is not a string")
		}

		lastPage, err := strconv.Atoi(e.Text)
		if err != nil {
			errChan <- errors.New("Value is NaN")
		}

		respChan <- lastPage
	})

	c.OnError(func(r *colly.Response, err error) {
		errChan <- err
	})

	c.Visit("https://vyvymanga.net/genre/" + genre)

	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(10 * time.Second):
		return 0, errors.New("timeout")
	case err := <-errChan:
		return 0, err
	}
}
