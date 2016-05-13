package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

var config Config

type SiteHandler struct {
	config    *SiteConfig
	forwarder *forward.Forwarder
	hostRegex *regexp.Regexp
}

func NewSiteHandler(config *SiteConfig) *SiteHandler {
	forwarder, err := forward.New()

	if err != nil {
		log.Fatal(err)
	}

	return &SiteHandler{
		config:    config,
		forwarder: forwarder,
		hostRegex: regexp.MustCompile(config.VirtualHost),
	}
}

func (s *SiteHandler) MatchHost(host string) bool {
	return s.hostRegex.Match([]byte(host))
}

func (s *SiteHandler) Handle(w http.ResponseWriter, request *http.Request) {
	request.URL = s.config.EndpointURL
	if s.config.HostHeader != "" {
		request.Host = s.config.HostHeader
	}

	bw := &bufferWriter{header: make(http.Header), buffer: &bytes.Buffer{}}

	s.forwarder.ServeHTTP(bw, request)

	body := s.readAndReplaceBody(bw)

	utils.CopyHeaders(w.Header(), bw.Header())

	headers := w.Header()
	if s.config.RewriteLocations {
		location := s.applyReplacements([]byte(headers.Get("Location")))
		headers.Set("Location", string(location))
	}
	headers.Set("Content-Length", strconv.Itoa(len(body)))
	headers.Del("Content-Encoding")
	w.WriteHeader(bw.code)
	w.Write(body)
}

func (s *SiteHandler) applyReplacements(body []byte) []byte {
	b := body
	for _, rule := range s.config.Replacements {
		b = rule.Source.ReplaceAll(b, []byte(rule.Target))
	}
	return b
}

func (s *SiteHandler) readAndReplaceBody(bw *bufferWriter) []byte {
	var reader io.Reader

	//Try to read it as gzip, if not leave it as is
	reader, err := gzip.NewReader(bw.buffer)
	if err != nil {
		reader = bw.buffer
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println(err)
	}

	return s.applyReplacements(body)
}

func main() {
	config = parseConfig()

	sites := []*SiteHandler{}

	for _, sc := range config.SiteConfigs {
		sites = append(sites, NewSiteHandler(sc))
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, s := range sites {
			if s.MatchHost(r.Host) {
				s.Handle(w, r)
				return
			}
		}
	})

	err := http.ListenAndServe(config.ListenURL, handler)
	if err != nil {
		log.Fatal(err)
	}
}
