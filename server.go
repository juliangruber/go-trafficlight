
// TODO timeout based on querystring

package main

import (
  "net/http"
  "net/url"
  "fmt"
  "time"
  "regexp"
  "io/ioutil"
)

var green, yellow, red = func () ([]byte, []byte, []byte) {
  green, err := ioutil.ReadFile("./images/green.gif")
  if err != nil { panic(err) }

  yellow, err := ioutil.ReadFile("./images/yellow.gif")
  if err != nil { panic(err) }

  red, err := ioutil.ReadFile("./images/red.gif")
  if err != nil { panic(err) }

  return green, yellow, red
}()

var colors = map[string][]byte{
  "green": green,
  "yellow": yellow,
  "red": red,
}

var reg, _ = regexp.Compile("^/http:\\/")

func testRegexp (r string, body []byte) (bool, error) {
  return regexp.Match(r, body)
}

func testNotRegexp (r string, body []byte) (bool, error) {
  matched, err := regexp.Match(r, body)
  return !matched, err
}

func getColor(url string, query url.Values, c chan string, e chan error) {
  start := time.Now()

  resp, err := http.Get(url)

  if err != nil {
    c <- "red"
    return
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)

  if err != nil {
    c <- "red"
    return
  }

  valid := true

  if _, found := query["regex"]; found {
    valid, err = testRegexp(query["regex"][0], body)
  } else if _, found := query["not-regex"]; found {
    valid, err = testNotRegexp(query["not-regex"][0], body)
  }

  if err != nil { e <- err }
  if !valid {
    c <- "red"
    return
  }

  if dt := time.Now().Sub(start); dt > time.Second {
    c <- "yellow"
    return
  } else {
    c <- "green"
    return
  }
}

func handler(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/favicon.ico" {
    fmt.Fprintf(w, "404")
    return
  }

  url := "http://" + reg.ReplaceAllString(r.URL.Path, "")
  query := r.URL.Query()

  color := make(chan string)
  err := make(chan error)

  go getColor(url, query, color, err)

  for {
    select {
      case c := <-color:
        w.Write(colors[c])
        return
      case e := <-err:
        fmt.Fprintf(w, e.Error())
        return
    }
  }
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}
