package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
)

const HeaderTargetDestination = "X-Checker-Destination"

func main() {
	var port int

	flag.IntVar(&port, "port", 80, "")
	flag.Parse()

	http.Handle("/", Api{})

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)

	if err != nil {
		log.Fatal(err)
	}
}

type Api struct {
}

func (o Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Method:     r.Method,
		RemoteAddr: r.RemoteAddr,
		Header:     r.Header,
	}

	resp.Header.Add(HeaderTargetDestination, "SUCCESS")

	bytes, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

type Response struct {
	Method     string
	RemoteAddr string
	Header     http.Header
}
