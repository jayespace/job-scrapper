package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id         string
	title      string
	location   string
	experience string
	company    string
}

// first get pages visit each page extract jobs and make it csv
func Scrap(term string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?searchType=search&searchword=" + term + "&exp_cd=1%2C2&exp_max=2&exp_none=y&panel_type=&search_optional_item=y&search_done=y&panel_count=y&preview=y&recruitPage=2&recruitSort=relation&recruitPageCount=40&inner_com_type=&show_applied=&quick_apply=&except_read=&ai_head_hunting="

	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)

	for i := 1; i <= totalPages; i++ {
		go getPage(baseURL, i, c)
	}

	for i := 1; i <= totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}

	go writeJobs(jobs)
	totalJobs := len(jobs)
	fmt.Printf("Done extracted %v jobs\n", totalJobs)
}

// Get total pages to look up
func getPages(baseURL string) int {
	pages := 0
	res, err := http.Get(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length() + 1
	})
	fmt.Println(pages)
	return pages
}

// Look through each page and get information
func getPage(baseURL string, pages int, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := baseURL + fmt.Sprintf("&recruitPage=%d", pages)

	res, err := http.Get(pageURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	searchJobs := doc.Find(".item_recruit")
	searchJobs.Each(func(i int, s *goquery.Selection) {
		go extractJob(s, c)
	})
	for i := 0; i < searchJobs.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

// From each page, find details of the jobs
func extractJob(s *goquery.Selection, c chan<- extractedJob) {
	id, _ := s.Attr("value")
	title := CleanString(s.Find(".job_tit a").Text())
	// title1 := s.Find("a").First()
	// title, _ := title1.Attr("title")

	company := CleanString(s.Find(".corp_name>a").Text())
	location := s.Find(".job_condition a").First().Text()
	experience := s.Find(".job_condition span:nth-child(2)").Text()
	c <- extractedJob{
		id:         id,
		title:      title,
		location:   location,
		experience: experience,
		company:    company,
	}
}

// Make the result into a csv file
func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	if err != nil {
		log.Fatal(err)
	}

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Company", "Title", "Location", "Experience", "Link"}
	err = w.Write(headers)
	if err != nil {
		log.Fatal(err)
	}

	for _, job := range jobs {
		jobURL := "https://www.saramin.co.kr/zf_user/jobs/view?rec_idx=" + job.id
		jobSlice := []string{job.company, job.title, job.location, job.experience, jobURL}
		err := w.Write(jobSlice)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
