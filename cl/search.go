package cl

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/SlyMarbo/rss"
)

const (
	jpgType       = "image/jpeg"
	searchCityStr = "%s/search/%s?format=rss&query=%s"
)

func init() {
	rss.CacheParsedItemIDs(false)
}

func Search(cityName, category, query string, distanceMI float64) ([]SearchResult, error) {
	city, err := GetCity(cityName)
	if err != nil {
		return nil, err
	}

	cities := append([]City{city}, city.CitiesWithin(distanceMI)...)
	results := make([]SearchResult, 0)

	var wg sync.WaitGroup
	var m sync.Mutex
	for _, city := range cities {
		wg.Add(1)
		go func() {
			defer wg.Done()
			posts, err := city.Search(category, query)
			if err != nil {
				log.Print(err)
				return
			}
			m.Lock()
			results = append(results, SearchResult{City: city, Posts: posts})
			m.Unlock()
		}()
	}
	wg.Wait()

	return results, nil
}

type Post struct {
	Title string
	URL   string
	Date  time.Time
	Image string
}

type SearchResult struct {
	City  City
	Posts []Post
}

func (c City) Search(category, query string) ([]Post, error) {
	feed, err := rss.Fetch(fmt.Sprintf(searchCityStr, c.URL, category, query))
	if err != nil {
		fmt.Println(fmt.Sprintf(searchCityStr, c.URL, category, query))
		return nil, err
	}

	results := make([]Post, 0)
	for _, item := range feed.Items {
		results = append(results, newPost(item))
	}
	return results, nil
}

func newPost(item *rss.Item) Post {
	post := Post{
		Title: item.Title,
		Date:  item.Date,
		URL:   item.Link,
	}
	for _, e := range item.Enclosures {
		if e.Type == jpgType {
			post.Image = e.Url
			break
		}
	}
	return post
}
