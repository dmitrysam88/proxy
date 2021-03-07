package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func getConfig() (int, map[string]string, error) {
	var config map[string]json.RawMessage
	var port int
	var routes map[string]string
	jsonFile, err := os.Open("config.json")

	defer jsonFile.Close()

	if err != nil {
		fmt.Println(err)
		return 0, routes, errors.New("File config.json not found")
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &config)

	json.Unmarshal(config["port"], &port)

	json.Unmarshal(config["routes"], &routes)

	return port, routes, nil
}

func getProxyMap(routes map[string]string) map[string]*httputil.ReverseProxy {
	result := make(map[string]*httputil.ReverseProxy)
	var target *url.URL
	for key, value := range routes {
		target = &url.URL{Scheme: "http", Host: value}
		result[key] = httputil.NewSingleHostReverseProxy(target)
	}
	return result
}

func main() {

	port, routes, err := getConfig()
	if err == nil {
		proxyMap := getProxyMap(routes)
		http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
			scheme := "http"
			if req.URL.Scheme == "https" {
				scheme = req.URL.Scheme
			}

			fmt.Printf("- %s://%s%s %s \n", scheme, req.Host, req.URL.Path, req.Method)
			currentProxy := proxyMap[req.Host]

			if currentProxy == nil {
				res.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(res, "Service %q not found", html.EscapeString(req.Host))
			} else {
				currentProxy.ServeHTTP(res, req)
			}
		})

		fmt.Printf("Proxy server run on http://localhost:%d/ \n", port)
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	} else {
		fmt.Println(err)
	}
}
