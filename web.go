package main

import(
  "net/http"
  "net/url"
  "log"
  "fmt"
  "time"
  "os"

  "code.google.com/p/go-html-transform/h5"
  "code.google.com/p/go.net/html"
)

type Aggregates map[string][]down

type down struct {
  Origin string
  Url    string
  Status int
}

func (d down) String() string {
  return fmt.Sprintf("%s: %d %s", d.Origin, d.Status, d.Url)
}

var (
  downs = make(chan down)
)

func main() {
  // start aggregate routine
  aggregates()

  // start server that triggers checks
  http.HandleFunc("/trigger", func(w http.ResponseWriter, _ *http.Request) {
    fmt.Fprintf(w, "ok")
    checkURL("http://freenerd.de/")
  })
  var port string
  if port = os.Getenv("PORT"); port == "" {
    port = "8080"
  }
  fmt.Printf("listening on port %s ...\n", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}

func aggregates() {
  go func() {
    for {
      collectAggregates()
      time.Sleep(1 * time.Minute)
    }
  }()
}

func collectAggregates() {
  ds := Aggregates{}

  loop: for {
    select {
    case d := <-downs:
      ds[d.Origin] = append(ds[d.Origin], d)
    default:
      // drained channel, continue with processing
      break loop
    }
  }

  if(os.Getenv("SENDGRID_USERNAME") != "" &&
       os.Getenv("SENDGRID_PASSWORD") != "" &&
       os.Getenv("EMAIL_RECIPIENT") != "") {
    emailAggregates(ds)
  } else {
    printAggregates(ds)
  }
}

func printAggregates(ds Aggregates) {
  for _, downs := range ds {
    for _, d := range downs {
      log.Println(d.String())
    }
  }
}

func emailAggregates(ds Aggregates) {
  for origin, downs := range ds {
    text := ""
    for _, d := range downs {
      text = fmt.Sprintf("%s\n%s", text, d.String())
    }

    v := url.Values{}
    v.Add("api_user", os.Getenv("SENDGRID_USERNAME"))
    v.Add("api_key", os.Getenv("SENDGRID_PASSWORD"))
    v.Add("to", os.Getenv("EMAIL_RECIPIENT"))
    v.Add("toname", os.Getenv("EMAIL_RECIPIENT"))
    v.Add("subject", origin)
    v.Add("text", text)
    v.Add("from", os.Getenv("EMAIL_RECIPIENT"))

    resp, err := http.PostForm("https://api.sendgrid.com/api/mail.send.json", v)
    log.Println("Email sent")
    log.Println(resp)
    log.Println(err)
  }
}

func checkURL(origin string) {
  // parse incoming url
  up, err := url.Parse(origin)
  if err != nil {
    log.Println(err)
    return
  }
  SCHEME := up.Scheme
  HOST := up.Host

  resp, err := http.Get(origin)
  if err != nil {
    log.Println(err)
    return
  }

  tree, err := h5.New(resp.Body)
  if err != nil {
    log.Println(err)
    return
  }

  tree.Walk(func(n *html.Node) {
    if n.Type == html.ElementNode && n.Data == "a" {
      for _, attr := range n.Attr {
        if attr.Key == "href" {
          u, err := url.Parse(attr.Val)
          if err != nil {
            log.Println(err)
            continue
          }

          if !u.IsAbs() {
            // if its relative, we assume it's relative from the current path
            u.Scheme = SCHEME
            u.Host = HOST
          }

          checkHead(origin, u.String())
        }
      }
    }
  })
}

func checkHead(origin, url string) {
  go func() {
    resp, err := http.Head(url)
    if err != nil {
      // assuming problems with network or malformed url
      // ignore
      return
    }

    if (resp.StatusCode >= 400) {
      // url is in trouble
      // some sites like amazon or google app engine don't like HEAD, let's retry with GET
      checkGet(origin, url)
    }
  }()
}

func checkGet(origin, url string) {
  go func() {
    resp, err := http.Get(url)
    if err != nil {
      // assuming problems with network or malformed url
      // ignore
      return
    }

    if (resp.StatusCode >= 400) {
      // url is down, down, down
      downs <- down{
        Origin: origin,
        Url:    url,
        Status: resp.StatusCode,
      }
    }
  }()
}
