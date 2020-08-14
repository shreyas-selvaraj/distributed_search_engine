package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	rake "github.com/afjoseph/RAKE.Go"
	"github.com/jackdanger/collectlinks"
)

// type Link struct {
// 	Link     string   `json:"link"`
// 	Keywords []string `json:"keywords"`
// }

// func GetESClient() (*elasticsearch.Client, error) {

// 	cfg := elasticsearch.Config{
// 		Addresses: []string{
// 			"http://localhost:9200",
// 			"http://localhost:9201",
// 		},
// 		// ...
// 	}
// 	es, err := elasticsearch.NewClient(cfg)

// 	fmt.Println("ES initialized...")

// 	return es, err

// }

// func usage() {
// 	fmt.Fprintf(os.Stderr, "usage: crawl http://example.com/path/file.html\n")
// 	flag.PrintDefaults()
// 	os.Exit(2)
// }

func handler(w http.ResponseWriter, r *http.Request) {
	args := []string{"https://www.nytimes.com/"}

	queue := make(chan string)
	filteredQueue := make(chan string)

	go func() { queue <- args[0] }()
	go filterQueue(queue, filteredQueue)

	// introduce a bool channel to synchronize execution of concurrently running crawlers
	done := make(chan bool)

	// pull from the filtered queue, add to the unfiltered queue
	for i := 0; i < 5; i++ {
		go func() {
			for uri := range filteredQueue {
				enqueue(uri, queue, w)
			}
			done <- true
		}()
	}
	<-done
}

func main() {
	log.Print("helloworld: starting server...")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

	// args := []string{"https://www.nytimes.com/"}

	// queue := make(chan string)
	// filteredQueue := make(chan string)

	// go func() { queue <- args[0] }()
	// go filterQueue(queue, filteredQueue)

	// // introduce a bool channel to synchronize execution of concurrently running crawlers
	// done := make(chan bool)

	// // pull from the filtered queue, add to the unfiltered queue
	// for i := 0; i < 5; i++ {
	// 	go func() {
	// 		for uri := range filteredQueue {
	// 			enqueue(uri, queue)
	// 		}
	// 		done <- true
	// 	}()
	// }
	// <-done
}

func filterQueue(in chan string, out chan string) {
	var seen = make(map[string]bool)
	for val := range in {
		if !seen[val] {
			seen[val] = true
			out <- val
		}
	}
}

func GetLatestBlogTitles(url string) (string, error) {

	// Get the HTML
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	// Convert HTML into goquery document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// Save each .post-title as a list
	titles := ""
	// doc.Find("meta").Each(func(i int, s *goquery.Selection) {
	// 	titles += "- " + s.Text() + "\n"
	// })
	titles = doc.Find("p").Text()
	return titles, nil
}

// func pushElastic(link string, keywords []string) {
// 	ctx := context.Background()
// 	esclient, err := GetESClient()
// 	if err != nil {
// 		fmt.Println("Error initializing : ", err)
// 		panic("Client fail ")
// 	}

// 	newLink := Link{
// 		Link:     link,
// 		Keywords: keywords,
// 	}

// 	dataJSON, err := json.Marshal(newLink)
// 	js := string(dataJSON)

// 	req := esapi.IndexRequest{
// 		Index:   "links",
// 		Body:    strings.NewReader(js),
// 		Refresh: "true",
// 	}

// 	ind, err := req.Do(ctx, esclient)

// 	if err != nil {
// 		panic(err)
// 	}

// 	print(ind)

// 	fmt.Println("[Elastic][InsertProduct]Insertion Successful")
// }

func enqueue(uri string, queue chan string, w http.ResponseWriter) {
	fmt.Println("fetching", uri)
	fmt.Fprintf(w, "fetching: "+uri)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)
	if err != nil {
		return
	}

	response, err := GetLatestBlogTitles(uri)
	if err != nil {
		log.Println(err)
	}
	//fmt.Println(resp)

	candidates := rake.RunRake(response)
	var candidate_keys []string
	i := 0

	for _, candidate := range candidates {
		if i > 5 {
			break
		}
		fmt.Printf("%s --> %f\n", candidate.Key, candidate.Value)
		candidate_keys = append(candidate_keys, candidate.Key)
		i++
	}

	fmt.Printf("\nsize: %d\n", len(candidates))

	//pushElastic(uri, candidate_keys)

	defer resp.Body.Close()

	links := collectlinks.All(resp.Body)

	for _, link := range links {
		absolute := fixUrl(link, uri)
		if uri != "" {
			go func() { queue <- absolute }()
		}
	}
}

func fixUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}
