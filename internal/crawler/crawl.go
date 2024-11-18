package crawler

import (
	"context"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

type Crawler struct {
	sema    chan struct{}
	wg      sync.WaitGroup
	timeout time.Duration
}

func NewCrawler(concurrency int, timeout time.Duration) *Crawler {
	return &Crawler{
		sema:    make(chan struct{}, concurrency),
		timeout: timeout,
	}
}
func (c *Crawler) CrawlPage(url string) (string, error) {
	c.sema <- struct{}{}
	c.wg.Add(1)

	result := make(chan string)
	url = convertURL(url)

	go func() {
		defer func() {
			<-c.sema
			c.wg.Done()
			close(result)
		}()

		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()

		var html string

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.InnerHTML("html", &html),
		)
		if err != nil {
			return
		}
		result <- html
	}()

	select {
	case <-time.After(c.timeout):
		return "", nil
	case html, ok := <-result:
		if !ok {
			return "", nil
		}
		return html, nil
	}
}
