package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var googleDomains = map[string]string{
	"com": "https://www.google.com/search?q=",
}

type Result struct {
	Rank int
	URL  string
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

func RandomUserAgent() string {
	rand.Seed(time.Now().Unix())
	randomNumber := rand.Int() % len(userAgents)

	return userAgents[randomNumber]
}

func buildGoogleUrls(searchTerm, googleDomainExtension string, pages int) ([]string, error) {
	scrape := []string{}
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if base, found := googleDomains[googleDomainExtension]; found {
		for i := 0; i < pages; i++ {
			scrapeURL := fmt.Sprint(base, searchTerm)
			scrape = append(scrape, scrapeURL)
		}
	} else {
		err := fmt.Errorf("Not supported domain extension")
		return nil, err
	}

	return scrape, nil
}

func GoogleScrape(searchTerm string, googleDomainExtension string, proxyString interface{}, pages, backoff int) ([]Result, error) {
	results := []Result{}
	resultCounter := 0
	googlePages, err := buildGoogleUrls(searchTerm, googleDomainExtension, pages)

	if err != nil {
		return nil, err
	}

	for _, page := range googlePages {
		res, err2 := scrapeClientRequest(page, proxyString)
		if err2 != nil {
			return nil, err2
		}
		data, err3 := googleResultParsing(res, resultCounter)
		if err3 != nil {
			return nil, err3
		}
		resultCounter += len(data)
		for _, result := range data {
			results = append(results, result)
		}
		time.Sleep(time.Duration(backoff) * time.Second)
	}

	return results, nil
}

func getScrapeClient(proxyString interface{}) *http.Client {
	switch v := proxyString.(type) {
	case string:
		proxyUrl, _ := url.Parse(v)
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	default:
		return &http.Client{}
	}

}

func googleResultParsing(response *http.Response, rank int) ([]Result, error) {
	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		return nil, err
	}

	results := []Result{}
	sel := doc.Find("div.g")
	rank++
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		link = strings.Trim(link, " ")

		if link != "" && link != "#" && !strings.HasPrefix(link, "/") {
			result := Result{
				rank,
				link,
			}
			results = append(results, result)
			rank++
		}
	}
	return results, err

}

func scrapeClientRequest(searchURL string, proxyString interface{}) (*http.Response, error) {
	baseClient := getScrapeClient(proxyString)
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", RandomUserAgent())

	res, err := baseClient.Do(req)
	if res.StatusCode != 200 {
		err := fmt.Errorf("scraper received a non 200 status code ")
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return res, nil
}

func main() {
	res, err := GoogleScrape("tayfun güler", "com", nil, 2, 5)
	if err == nil {
		for _, res := range res {
			fmt.Println(res)
		}
	} else {
		fmt.Println(err)
	}
}
