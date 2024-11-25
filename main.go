package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"golang.design/x/clipboard"
	"golang.org/x/net/html"
)

var (
	dataFileName  = "data/data.json"
	urlRegex      = `^(http|https)?://(.*)?(/.*)?$`
	throttleLimit = 10
)

type Url struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

var datas []*Url

func SaveData(filename string, url []*Url) error {
	data, err := json.MarshalIndent(url, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, os.ModePerm)
}

func ReadData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &datas)
}

func CheckData(url string) bool {
	for _, x := range datas {
		if x.URL == url {
			return true
		}
	}
	return false
}

func Notify(title, msg string) error {
	return beeep.Notify(title, msg, "assets/information.png")
}

func Alert(title, msg string) error {
	return beeep.Alert(title, msg, "assets/warning.png")
}

func GetTitle(ctx context.Context, url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("url not reached")
	}
	return ParseTitle(res.Body)
}

func GetTitleWg(ctx context.Context, url *Url, throttle chan struct{}, result chan<- *Url, wg *sync.WaitGroup) {
	defer wg.Done()
	throttle <- struct{}{}
	defer func() { <-throttle }()
	res, err := http.Get(url.URL)
	if err != nil {
		result <- nil
		return
	}
	defer res.Body.Close()
	title, err := ParseTitle(res.Body)
	if err != nil {
		result <- nil
		return
	}
	url.Title = title
	result <- url
}

func ParseTitle(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}
	var title string
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "title" {
			if node.FirstChild != nil {
				title = node.FirstChild.Data
			}
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(doc)
	return title, nil
}

func main() {
	if err := ReadData(dataFileName); err != nil {
		log.Println(err)
	}
	if err := clipboard.Init(); err != nil {
		log.Fatal(err)
	}
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM)
	cbCh := clipboard.Watch(context.TODO(), clipboard.FmtText)
	if len(datas) > 0 {
		var wg sync.WaitGroup
		results := make(chan *Url, len(datas))
		throttle := make(chan struct{}, throttleLimit)
		log.Println("wg start check url title")
		for _, item := range datas {
			if item.Title == "" {
				wg.Add(1)
				go GetTitleWg(context.TODO(), item, throttle, results, &wg)
			}
		}
		go func() {
			wg.Wait()
			close(results)
			close(throttle)
			log.Println("wg closed")
		}()
		for result := range results {
			for _, d := range datas {
				if result != nil && d.URL == result.URL {
					d.Title = result.Title
					log.Println("get title url:", d.URL, d.Title)
					break
				}
			}
		}
		SaveData(dataFileName, datas)
	}
	regex := regexp.MustCompile(urlRegex)
	log.Println("urlclip start watching...")
loop:
	for {
		select {
		case text := <-cbCh:
			url := regex.FindString(string(text))
			if url != "" {
				title, _ := GetTitle(context.TODO(), url)
				log.Println("saved url:", url, title)
				if !CheckData(url) {
					datas = append(datas, &Url{
						URL:       url,
						Title:     title,
						CreatedAt: time.Now(),
					})
					if SaveData(dataFileName, datas) == nil {
						Notify("url saved", url)
					}
				} else {
					Alert("url already exist", url)
				}
			}
		case <-quitCh:
			log.Println("gracefull termination")
			close(quitCh)
			break loop
		}
	}
}
