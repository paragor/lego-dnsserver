package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/paragor/lego-dnsserver/pkg"
	"io"
	"log"
	"net/http"
)

type legoRequest struct {
	Fqdn  string `json:"fqdn"`
	Value string `json:"value"`
}

func parseLegoRequest(body io.Reader) (*legoRequest, error) {
	var legoR legoRequest
	err := json.NewDecoder(body).Decode(&legoR)
	if err != nil {
		return nil, err
	}
	if legoR.Fqdn == "" || legoR.Value == "" {
		return nil, fmt.Errorf("wrong lego request: fqdn or value is empty")
	}

	return &legoR, nil
}

func main() {
	var listenHttp string
	var listenDns string
	flag.StringVar(
		&listenHttp,
		"listen-http",
		"127.0.0.1:18888",
		"Listen addr for serve lego httpreq httpreq https://go-acme.github.io/lego/dns/httpreq/",
	)
	flag.StringVar(
		&listenDns,
		"listen-dns",
		"127.0.0.1:5352",
		"Listen addr for serve dns records",
	)
	flag.Parse()

	dnsProvider, err := pkg.NewDNSProvider(listenDns)
	if err != nil {
		panic(err)
	}

	handlerFunc := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			http.Error(writer, "only POST allowed", http.StatusMethodNotAllowed)
			log.Println("only POST allowed")
			return
		}
		defer request.Body.Close()
		if request.RequestURI == "/present" {
			//nolint:govet
			legoR, err := parseLegoRequest(request.Body)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				log.Println(err.Error())
				return
			}
			err = dnsProvider.Present(legoR.Fqdn, legoR.Value)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				log.Println(err.Error())
				return
			}
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte("OK"))
			return
		}

		if request.RequestURI == "/cleanup" {
			err = dnsProvider.CleanUp()
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				log.Println(err.Error())
				return
			}
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte("OK"))
			return
		}

		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte("unknown path"))
		log.Println("unknown path")
		return
	})
	log.Println("starting...")
	err = http.ListenAndServe(listenHttp, LogMiddleware(handlerFunc))
	if err != nil {
		panic(err)
	}
}

func LogMiddleware(origin http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.Method, request.RequestURI, request.RemoteAddr)
		origin.ServeHTTP(writer, request)
	})
}
