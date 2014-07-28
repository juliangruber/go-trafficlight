// TODO timeout based on querystring

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var reg = regexp.MustCompile("^/http:\\/")

func getColor(url string, query url.Values) (string, error) {
	start := time.Now()
	resp, err := http.Get(url)

	if err != nil {
		return "red", nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "red", nil
	}

	valid := true

	if _, found := query["regex"]; found {
		valid, err = regexp.Match(query["regex"][0], body)
	} else if _, found := query["not-regex"]; found {
		valid, err = regexp.Match(query["not-regex"][0], body)
		valid = !valid
	}

	if err != nil {
		return "", err
	}

	if !valid {
		return "red", nil
	}

	if dt := time.Now().Sub(start); dt > time.Second {
		return "yellow", nil
	} else {
		return "green", nil
	}
}

func newHandler() http.HandlerFunc {
	colors := make(map[string][]byte)
	colors["green"], _ = ioutil.ReadFile("images/green.gif")
	colors["yellow"], _ = ioutil.ReadFile("images/yellow.gif")
	colors["red"], _ = ioutil.ReadFile("images/red.gif")

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			fmt.Fprintf(w, "404")
			return
		}

		url := "http://" + reg.ReplaceAllString(r.URL.Path, "")
		query := r.URL.Query()

		color, err := getColor(url, query)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			w.Write(colors[color])
		}
	}
}

func main() {
	http.HandleFunc("/", newHandler())
	http.ListenAndServe(":8080", nil)
}
