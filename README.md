rewriter-proxy
=============

Small proxy written in go that accepts series of host settings each one with:

* Its own set of body replacement rules(regexes)
* The endpoint url where it has to proxy the connection matching that host
* If in addition to the body it also has to rewrite Location response headers
* The Host header to send to the site

To install:

* Clone into your `$GOPATH
* `go get .`
* `go build .`
