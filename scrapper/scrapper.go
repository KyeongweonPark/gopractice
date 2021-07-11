package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	link        string
	title       string
	companyName string
	location    string
	summary     string
}

var baseDomain string = "http://kr.indeed.com"
var baseURL string

func Scrape(term string) {
	baseURL = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"
	c := make(chan []extractedJob)
	var jobs []extractedJob
	totalPages := getPages()
	// totalPages = 1

	for i := 0; i < totalPages; i++ {
		go getPage(i, c)

	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")
	// fmt.Println(doc.Html())

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})
	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs

}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	link, _ := card.Attr("href")
	link = baseDomain + link
	title := CleanString(card.Find(".jobTitle>span").Text())
	companyName := CleanString(card.Find(".companyName").Text())
	companyLoc := CleanString(card.Find(".companyLocation").Text())
	summary := CleanString(card.Find(".job-snippet").Text())
	c <- extractedJob{
		link:        link,
		title:       title,
		companyName: companyName,
		location:    companyLoc,
		summary:     summary}
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	// fmt.Println(doc)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	return pages
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link", "Title", "CompanyName", "Location", "Summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.link, job.title, job.companyName, job.location, job.summary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("")
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
