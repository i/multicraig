package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	http.HandleFunc("/cities", func(w http.ResponseWriter, r *http.Request) {
		cities, err := GetCities()
		if err != nil {
			fmt.Fprint(w, err)
		}
		fmt.Fprint(w, cities)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		city := r.FormValue("city")
		query := r.FormValue("query")
		res, err := SearchCity(city, query)
		if err != nil {
			fmt.Fprintf(w, "Sorry, bud")
			return
		}

		fmt.Fprintf(w, res)
	})

	GetCities()
	http.ListenAndServe("localhost:3000", nil)
}

const searchCityStr = "http://%s.craigslist.org/search/sss?format=rss&query=%s"

func SearchCity(city, query string) (string, error) {
	res, err := http.Get(fmt.Sprintf(searchCityStr, city, query))
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status from CL rss: %d", res.Status)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(buf), err
}

const cityURL = "http://www.craigslist.org/about/sites"

type city struct {
	name   string
	region string
	url    string
}

var cityRE = regexp.MustCompile(`<li>.*http://([a-z]*)\.craigslist.*</a>`)
var regionRE = regexp.MustCompile(`<h4>(.*)</h4>`)

func GetCities() ([]city, error) {
	res, err := http.Get(cityURL)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from CL rss: %d", res.Status)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var cities []city
	var region string
	for _, line := range strings.Split(string(buf), "\n") {
		m := regionRE.FindSubmatch([]byte(line))
		if len(m) > 1 {
			region = string(m[1])
			continue
		}

		m = cityRE.FindSubmatch([]byte(line))
		if len(m) > 1 {
			c := city{
				name:   string(m[1]),
				region: region,
			}
			cities = append(cities, c)
		}
	}

	return cities, nil
}
