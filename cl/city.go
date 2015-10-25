package cl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"multicraig/storage"

	"github.com/SlyMarbo/rss"
	"github.com/i/jdog"
)

const (
	kmtomiles       = float64(0.621371192)
	earthRadius     = float64(6371)
	clCityURL       = "http://www.craigslist.org/about/sites"
	searchCityStr   = "http://%s.craigslist.org/search/%s?format=rss&query=%s"
	cityLocationURL = "https://maps.googleapis.com/maps/api/geocode/json?address=%s,+%s&key=%s"
)

var (
	//cityRE   = regexp.MustCompile(`<li>.*http://.*"([a-z]*)\.craigslist.*">(.*)</a>`)
	//cityRE   = regexp.MustCompile(`^<li><a href="(.*)".*>(.*)</a></li>$`)
	cityRE   = regexp.MustCompile(`^<li><a href="(.*)">([a-zA-Z|\s]+)</a></li>$`)
	regionRE = regexp.MustCompile(`<h4>(.*)</h4>`)
	wwwRE    = regexp.MustCompile(`www.craigslist.org`)
	gapikey  = os.Getenv("MAPSAPIKEY")

	cache = storage.NewStore()
)

type City struct {
	Name   string
	Region string
	URL    string
	Lat    float64
	Lng    float64
}

func GetCity(name string) (City, error) {
	cities, err := GetCities()
	if err != nil {
		return City{}, err
	}
	for _, c := range cities {
		if c.Name == name {
			return c, nil
		}
	}
	return City{}, fmt.Errorf("city not found")
}

func GetCities() ([]City, error) {
	c, err := cache.Get("cities")
	if err != nil {
		log.Print(err)
	}
	if cities, ok := c.([]City); ok {
		if len(cities) > 0 {
			return cities, nil
		}
	}

	var cities []City
	defer func() { cache.Set("cities", cities) }()

	res, err := http.Get(clCityURL)
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

	var region string
	for _, line := range strings.Split(string(buf), "\n") {
		m := regionRE.FindStringSubmatch(line)
		if len(m) > 1 {
			region = m[1]
			continue
		}

		m = cityRE.FindStringSubmatch(line)
		if len(m) > 1 {
			if wwwRE.MatchString(m[1]) {
				continue
			}

			c := City{
				Region: region,
				Name:   m[2],
				URL:    m[1],
			}
			cities = append(cities, c)
		}
	}

	var wg sync.WaitGroup
	for i, c := range cities {
		//TODO -- remove when real
		if i == 5 {
			break
		}
		wg.Add(1)
		go func(i int, c City) {
			defer wg.Done()
			var err error
			c.Lat, c.Lng, err = c.GetLocation()
			if err != nil {
				log.Print(err)
				return
			}
			cities[i] = c
		}(i, c)
	}
	wg.Wait()

	return cities, nil
}

type Post struct {
	Title string
	URL   string
	Image string
}

func SearchCity(city, category, query string) (string, error) {
	// default category to for sale
	if category == "" {
		category = "sss"
	}

	feed, err := rss.Fetch(fmt.Sprintf(searchCityStr, city, query))
	if err != nil {
		return "", err
	}

	for _, item := range feed.Items {
		fmt.Println(item)
	}
	return "", nil
}

func (c City) GetLocation() (lat, lng float64, err error) {
	if c.Lat != 0 && c.Lng != 0 {
		return c.Lat, c.Lng, nil
	}

	res, err := http.Get(fmt.Sprintf(cityLocationURL, c.Name, c.Region, gapikey))
	defer res.Body.Close()
	if err != nil {
		return 0, 0, err
	}
	if res.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("bad status code from google: %d", res.StatusCode)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, 0, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return 0, 0, err
	}

	v, err := jdog.Get(m, "results[0].geometry.location.lat")
	if err != nil {
		return 0, 0, fmt.Errorf("lat/lng not found for %s", c.Name)
	}
	lat = v.(float64)

	v, err = jdog.Get(m, "results[0].geometry.location.lng")
	if err != nil {
		return 0, 0, fmt.Errorf("lat/lng not found for %s", c.Name)
	}
	lng = v.(float64)

	return lat, lng, nil
}

func (c City) CitiesWithin(mi float64) []City {
	cities, err := GetCities()
	if err != nil {
		log.Printf("error fetching cities: %v", err)
		return nil
	}
	var closeCities []City
	for _, c2 := range cities {
		if c2.Name == c.Name {
			continue
		}
		dst := c.distanceToCity(c2)
		if dst == -1 {
			continue
		}
		if dst < mi {
			closeCities = append(closeCities, c2)
		}
	}
	return closeCities
}

func (c City) distanceToCity(dst City) float64 {
	if c.Lat == 0 || c.Lng == 0 || dst.Lat == 0 || dst.Lng == 0 {
		return -1
	}

	km := haversine(c.Lng, c.Lat, dst.Lng, dst.Lat)
	return km * kmtomiles
}

func haversine(lonFrom float64, latFrom float64, lonTo float64, latTo float64) (distance float64) {
	var deltaLat = (latTo - latFrom) * (math.Pi / 180)
	var deltaLon = (lonTo - lonFrom) * (math.Pi / 180)

	var a = math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latFrom*(math.Pi/180))*math.Cos(latTo*(math.Pi/180))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance = earthRadius * c

	return
}
