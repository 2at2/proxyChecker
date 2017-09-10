package checker

import (
	"encoding/json"
	"errors"
	"github.com/mono83/slf/wd"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const HeaderForwardedFor = "X-Forwarded-For"
const HeaderTargetDestination = "X-Checker-Destination"

type Module struct {
	list         map[string]string
	target       *url.URL
	log          wd.Watchdog
	maxQueueLoad int
}

func NewModule(target string, list map[string]string, watchdog wd.Watchdog) (*Module, error) {
	if list == nil || len(list) == 0 {
		return nil, errors.New("empty list of proxies")
	}

	uri, err := url.Parse(target)

	if err != nil {
		return nil, err
	}

	return &Module{
		list:         list,
		target:       uri,
		log:          watchdog,
		maxQueueLoad: 100,
	}, nil
}

func (m Module) Process() ([]string, error) {
	req := make(chan bool, m.maxQueueLoad)
	out := make(chan Result)

	var urls []*url.URL

	// Preparing urls list
	for _, proxy := range m.list {
		uri, err := url.Parse(proxy)

		if err != nil {
			m.log.Error("Unable to parse proxy :proxy - :err", wd.StringParam("proxy", proxy), wd.ErrParam(err))
			return nil, err
		}

		urls = append(urls, uri)
	}

	// Sending requests
	go func(urls []*url.URL) {
		for _, uri := range urls {
			req <- true

			go func(uri *url.URL) {
				m.log.Debug("Sending :url", wd.StringParam("url", uri.String()))

				client := &http.Client{
					Transport: &http.Transport{
						Proxy:                 http.ProxyURL(uri),
						DisableKeepAlives:     true,
						TLSHandshakeTimeout:   10 * time.Second,
						ExpectContinueTimeout: 1 * time.Second,
						IdleConnTimeout:       10 * time.Second,
					},
				}

				request := &http.Request{
					Method: "GET",
					URL:    m.target,
				}

				resp, err := client.Do(request)

				out <- Result{
					Proxy:    uri.String(),
					Response: resp,
					Error:    err,
				}

				m.log.Debug("Response :url", wd.StringParam("url", uri.String()))

				<-req
			}(uri)
		}
	}(urls)

	var goodProxies []string

	// Reading result
	for i := 0; i < len(m.list); i++ {
		result := <-out

		log := m.log.WithParams(wd.StringParam("url", result.Proxy))

		log.Info("Proxy :url")

		if result.Error != nil {
			log.Error("Error request - :err", wd.ErrParam(result.Error))
		} else if result.Response.StatusCode != 200 {
			log.Error("Error http code - :code", wd.IntParam("code", result.Response.StatusCode))
		} else {
			body, err := ioutil.ReadAll(result.Response.Body)

			if err != nil {
				return nil, err
			}

			defer result.Response.Body.Close()

			if !strings.Contains(string(body), HeaderTargetDestination) {
				log.Error("Error proxy :url - no destination")
				continue
			}

			var js ResultBody

			err = json.Unmarshal(body, &js)

			if err != nil {
				log.Error("Unable to unmarshal body from :url - :err", wd.ErrParam(err))
				return nil, err
			}

			if _, ok := js.Header["X-Checker-Destination"]; !ok {
				log.Error("Error - no destination")
			} else {
				forwardedFor, ok := js.Header[HeaderForwardedFor]

				if ok {
					log.Warning("Not anonymous :url")

					for _, val := range forwardedFor {
						log.Warning("Forwarded for :name", wd.NameParam(val))
					}
				} else {
					log.Info("Good proxy :url")
					goodProxies = append(goodProxies, result.Proxy)
				}
			}
		}
	}

	return goodProxies, nil
}

type Result struct {
	Proxy    string
	Response *http.Response
	Error    error
}

type ResultBody struct {
	Method     string              `json:"Method"`
	RemoteAddr string              `json:"RemoteAddr"`
	Header     map[string][]string `json:"Header"`
}
