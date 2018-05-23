package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"io/ioutil"
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

func TestRegistry(t *testing.T) {
	zkServers := []string{"47.104.11.15:2181"}

	data, err := ioutil.ReadFile("./httpapi.json")
	if err != nil {
		t.Fatal(err)
	}

	urlRegistry, err := libkv.NewStore(store.ZK, zkServers, nil)
	if err != nil {
		t.Fatal(err)
	}

	key := fmt.Sprintf("%s/%s/%s", "navi/service",
		"nlp", "prod")
	err = urlRegistry.Put(key, data, nil)
	if err != nil {
		t.Fatal(err)
	}

	key = fmt.Sprintf("%s/%s/%s", "navi/service",
		"nlp", "dev")
	err = urlRegistry.Put(key, data, nil)
	if err != nil {
		t.Fatal(err)
	}

	urlRegistry.Close()
}
