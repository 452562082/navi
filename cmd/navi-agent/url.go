package main

type UrlInfo struct {
	ApiUrls []ApiUrl `json:"api_urls"`
}

type ApiUrl struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}
