package main

import (
	"net/http"
	"net/url"
	"sync"

	"code.google.com/p/go-html-transform/h5"
	"code.google.com/p/go.net/html"
)

func checkURL(u string, downChan chan Down) error {
	// parse incoming url
	up, err := url.Parse(u)
	if err != nil {
		return err
	}
	scheme := up.Scheme
	host := up.Host

	// fetch html document
	resp, err := http.Get(u)
	if err != nil {
		return err
	}

	// parse html document into tree
	tree, err := h5.New(resp.Body)
	if err != nil {
		return err
	}

	// walk tree and check all <a href="">
	var wg sync.WaitGroup
	tree.Walk(func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href, err := url.Parse(attr.Val)
					if err != nil {
						// malformed url, ignore
						continue
					}

					// make relative URLs absolute
					if !href.IsAbs() {
						href.Scheme = scheme
						href.Host = host
					}

					wg.Add(1)
					go func(origin, url string) {
						defer wg.Done()
						checkHead(origin, url, downChan)
					}(u, href.String())
				}
			}
		}
	})

	wg.Wait()
	return nil
}

func checkHead(origin, url string, downChan chan Down) {
	resp, err := http.Head(url)
	if err != nil {
		// assuming problems with network or malformed url
		// ignore
		return
	}

	if resp.StatusCode >= 400 {
		// url is in trouble
		// some sites like amazon or google app engine don't like HEAD, let's retry with GET
		checkGet(origin, url, downChan)
	}
}

func checkGet(origin, url string, downChan chan Down) {
	resp, err := http.Get(url)
	if err != nil {
		// assuming problems with network or malformed url
		// ignore
		return
	}

	if resp.StatusCode >= 400 {
		// url is down, down, down
		downChan <- Down{
			Origin: origin,
			Url:    url,
			Status: resp.StatusCode,
		}
	}
}
