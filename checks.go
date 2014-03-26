package main

import (
	"net/http"
	"net/url"
	"sync"

	"code.google.com/p/go-html-transform/h5"
	"code.google.com/p/go.net/html"
)

// downloads a document from a url
// parses it as HTML
// walks the DOM and looks for links (<a href="">)
// for each link, emits a go routine that checks its status code
func checkURL(u string, checked Checked) error {
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

	// walk tree and collect all <a href="">
	hrefs := []string{}
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

					hrefs = append(hrefs, href.String())
				}
			}
		}
	})

	// check each href
	downs := make(Downs, len(hrefs))
	var wg sync.WaitGroup
	for _, href := range hrefs {
		wg.Add(1)
		go func(u, href string) {
			defer wg.Done()
			checkHead(u, href, downs)
		}(u, href)
	}

	// wait until all checks are done
	wg.Wait()
	checked <- downs
	return nil
}

func checkHead(origin, href string, downs Downs) {
	resp, err := http.Head(href)
	if err != nil {
		// assuming problems with network or malformed url
		// ignore
		return
	}

	if resp.StatusCode >= 400 {
		// url is in trouble
		// some sites like amazon or google app engine don't like HEAD, let's retry with GET
		checkGet(origin, href, downs)
	}
}

func checkGet(origin, href string, downs Downs) {
	resp, err := http.Get(href)
	if err != nil {
		// assuming problems with network or malformed url
		// ignore
		return
	}

	if resp.StatusCode >= 400 {
		// url is down, down, down
		downs <- Down{
			Origin: origin,
			Url:    href,
			Status: resp.StatusCode,
		}
	}
}
