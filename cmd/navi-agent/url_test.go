package main

import (
	"encoding/json"
	"testing"
)

func TestJson(t *testing.T) {
	var urlInfo UrlInfo

	urlInfo.ApiUrls = append(urlInfo.ApiUrls, ApiUrl{URL: "/hello", Method: "get"})
	urlInfo.ApiUrls = append(urlInfo.ApiUrls, ApiUrl{URL: "/test/v1/hello", Method: "get"})

	data, err := json.Marshal(urlInfo)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))

	var ui UrlInfo
	err = json.Unmarshal(data, &ui)
	if err != nil {
		t.Fatal(err)
	}
}
