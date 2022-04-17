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

	dnsServer, err := pkg.NewDNSServer(listenDns)
	if err != nil {
		panic(err)
	}

	serverMux := http.NewServeMux()
	serverMux.Handle("/present",
		PostOnlyMiddleware(
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				//nolint:govet
				legoR, err := parseLegoRequest(request.Body)
				if err != nil {
					msg := "cant parse logo request: " + err.Error()
					log.Println(msg)
					http.Error(writer, msg, http.StatusBadRequest)
					return
				}
				defer request.Body.Close()
				err = dnsServer.Present(legoR.Fqdn, legoR.Value)
				if err != nil {
					msg := "cant start dns server: " + err.Error()
					http.Error(writer, msg, http.StatusBadRequest)
					log.Println(msg)
					return
				}
				writer.WriteHeader(200)
				_, _ = writer.Write([]byte("OK"))
			}),
		),
	)

	serverMux.Handle("/cleanup",
		PostOnlyMiddleware(
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				err = dnsServer.CleanUp()
				if err != nil {
					http.Error(writer, err.Error(), http.StatusBadRequest)
					log.Println(err.Error())
					return
				}
				writer.WriteHeader(200)
				_, _ = writer.Write([]byte("OK"))
				return
			}),
		),
	)
	log.Println("starting...")
	err = http.ListenAndServe(listenHttp, LogMiddleware(serverMux))
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

func PostOnlyMiddleware(origin http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			http.Error(writer, "only POST allowed", http.StatusMethodNotAllowed)
			log.Println("405. only POST allowed")
			return
		}
		origin.ServeHTTP(writer, request)
	})
}
