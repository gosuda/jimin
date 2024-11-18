package convert

import (
	"net/url"
	"path"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

var htmlSanitizerPolicy = bluemonday.UGCPolicy().
	AddSpaceWhenStrippingTag(true).
	AllowRelativeURLs(true).
	AllowStyles().Globally()

func CleanHTML(html string) string {
	return htmlSanitizerPolicy.Sanitize(html)
}

func convertURL(target, reference string) string {
	u, err := url.Parse(target)
	if err != nil {
		return target
	}

	if u.Host == "" {
		cu, err := url.Parse(reference)
		if err != nil {
			return target
		}
		if strings.HasPrefix(u.Path, "/") {
			cu.Path = u.Path
		} else {
			cu.Path = path.Join(cu.Path, u.Path)
		}
		return cu.String()
	}

	return u.String()
}

func ConvertHTMLToMarkdown(html string, curl string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	if doc.Find("main").Length() > 0 {
		doc.Find("main").Each(func(i int, s *goquery.Selection) {
			mainHtml, err := s.Html()
			if err != nil {
				return
			}
			html = mainHtml
		})
	}

	cleaned := CleanHTML(html)

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(cleaned))
	if err != nil {
		return "", err
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			s.SetAttr("src", convertURL(src, curl))
		}
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			s.SetAttr("href", convertURL(href, curl))
		}
	})

	cleaned, err = doc.Html()
	if err != nil {
		return "", err
	}

	converted, err := htmltomarkdown.ConvertString(cleaned)
	if err != nil {
		return "", err
	}

	converted = strings.ReplaceAll(converted, "\r\n", "\n")
	converted = strings.ReplaceAll(converted, "\n\n\n", "\n\n")
	converted = strings.ReplaceAll(converted, "\u200b", "")

	for strings.Contains(converted, " \n") {
		converted = strings.ReplaceAll(converted, " \n", "\n")
	}

	for strings.Contains(converted, "\n\n\n") {
		converted = strings.ReplaceAll(converted, "\n\n\n", "\n\n")
	}

	return converted, nil
}
