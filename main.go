package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"multicraig/cl"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	http.HandleFunc("/cities", func(w http.ResponseWriter, r *http.Request) {
		cities, err := cl.GetCities()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		buf, err := json.Marshal(cities)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		w.Write(buf)
	})

	http.HandleFunc("/closeCities", func(w http.ResponseWriter, r *http.Request) {
		distanceMi, err := strconv.ParseFloat(r.FormValue("distanceMi"), 64)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		city, err := cl.GetCity(r.FormValue("city"))
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		cities := city.CitiesWithin(distanceMi)
		buf, err := json.Marshal(cities)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		w.Write(buf)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		city := r.FormValue("city")
		query := r.FormValue("query")
		category := r.FormValue("category")
		if category == "" {
			category = "sss"
		}
		distanceMi, err := strconv.ParseFloat(r.FormValue("distanceMi"), 64)
		if err != nil {
			distanceMi = 0
		}

		results, err := cl.Search(city, category, query, distanceMi)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		buf, err := json.Marshal(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err.Error())
		}

		w.Write(buf)
	})

	http.ListenAndServe("localhost:3000", nil)
}
