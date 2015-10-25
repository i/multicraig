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
		miles, err := strconv.ParseFloat(r.FormValue("miles"), 64)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		city, err := cl.GetCity(r.FormValue("city"))
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		cities := city.CitiesWithin(miles)
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
		res, err := cl.SearchCity(city, category, query)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		fmt.Fprintf(w, res)
	})

	http.ListenAndServe("localhost:3000", nil)
}
