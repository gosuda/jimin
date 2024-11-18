package crawler

import "net/url"

func convertURL(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}

	switch u.Host {
	case "blog.naver.com":
		u.Host = "m.blog.naver.com"
		return u.String()
	}

	return path
}
