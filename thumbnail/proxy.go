package thumbnail

import (
	"fmt"
	"github.com/die-net/fotomat/format"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const (
	DefaultAccept    = "image/jpeg,image/*;q=0.6"
	DefaultServer    = "Fotomat"
	DefaultUserAgent = "Fotomat (http://fotomat.org)"
)

// Proxy represents an HTTP proxy that can optionally run its contents
// through Thumbnail.
type Proxy struct {
	Director  func(*http.Request) (Options, int)
	Client    *http.Client
	Accept    string
	Server    string
	UserAgent string
	pool      *Pool
	active    chan bool
}

func NewProxy(director func(*http.Request) (Options, int), pool *Pool, maxActive int, client *http.Client) *Proxy {
	if director == nil || pool == nil || client == nil || maxActive <= 0 {
		return nil
	}

	p := &Proxy{
		Director:  director,
		Client:    client,
		Accept:    DefaultAccept,
		Server:    DefaultServer,
		UserAgent: DefaultUserAgent,
		pool:      pool,
		active:    make(chan bool, maxActive),
	}

	for i := 0; i < maxActive; i++ {
		p.active <- true
	}

	return p
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, or *http.Request) {
	aborted := w.(http.CloseNotifier).CloseNotify()

	w.Header().Set("Server", p.Server)

	if or.Method != "GET" && or.Method != "HEAD" {
		proxyError(w, nil, http.StatusMethodNotAllowed)
		return
	}

	options, status := p.Director(or)
	if status != 0 {
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Wait for our turn to fetch and hold the original image.
	<-p.active

	if hasAborted(aborted) {
		p.active <- true // Release semaphore ASAP.
		proxyError(w, ErrAborted, 0)
		return
	}

	orig, header, status, err := p.get(or.URL.String(), or.Header)
	if err != nil || (status != http.StatusOK && status != http.StatusNotModified) {
		p.active <- true // Release semaphore ASAP.
		proxyError(w, err, status)
		return
	}

	copyHeaders(header, w.Header(), []string{"Age", "Cache-Control", "Etag", "Expires", "Last-Modified"})

	if status == http.StatusNotModified || isNotModified(or.Header, header) {
		p.active <- true // Release semaphore ASAP.
		w.WriteHeader(http.StatusNotModified)
		return
	}

	thumb, err := p.pool.Thumbnail(orig, options, aborted)
	orig = nil       // Free up image memory ASAP.
	p.active <- true // Release semaphore ASAP.

	if err != nil {
		proxyError(w, err, 0)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(thumb)))
	w.Write(thumb)
}

func (p *Proxy) get(url string, header http.Header) ([]byte, http.Header, int, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	// Pass some headers on to upstream.
	r.Header.Set("Accept", p.Accept)
	r.Header.Set("User-Agent", p.UserAgent)
	copyHeaders(header, r.Header, []string{"Cache-Control", "If-Modified-Since", "If-None-Match"})

	resp, err := p.Client.Do(r)
	if err != nil {
		return nil, nil, 0, err
	}

	orig, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return orig, resp.Header, resp.StatusCode, err
}

func (p *Proxy) Close() {
	close(p.active)
	p.pool.Close()
	*p = Proxy{}
}

func copyHeaders(src http.Header, dest http.Header, keys []string) {
	for _, key := range keys {
		if value, ok := src[key]; ok {
			dest[key] = value
		}
	}
}

func isNotModified(req http.Header, resp http.Header) bool {
	etag := resp.Get("Etag")
	match := req.Get("If-None-Match")
	// TODO: Support the multi-valued form of If-None-Match.
	if etag != "" && (match == etag || match == "*") {
		return true
	}

	// TODO: String compare of time is sub-optimal.
	lastMod := resp.Get("Last-Modified")
	since := req.Get("If-Modified-Since")
	return lastMod != "" && since == lastMod
}

func proxyError(w http.ResponseWriter, err error, status int) {
	switch status {
	case http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusRequestTimeout,
		http.StatusGone:
		err = nil
	case 0:
		switch err {
		case format.ErrUnknownFormat, ErrTooSmall:
			status = http.StatusUnsupportedMediaType
		case ErrTooBig:
			status = http.StatusRequestEntityTooLarge
		default:
			if isTimeout(err) {
				err = nil
				status = http.StatusGatewayTimeout
			} else {
				status = http.StatusInternalServerError
			}
		}
	default:
		err = fmt.Errorf("Proxy received %d %s", status, http.StatusText(status))
		status = http.StatusBadGateway
	}

	if err == nil {
		err = fmt.Errorf(http.StatusText(status))
	}

	http.Error(w, err.Error(), status)
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case net.Error:
		return err.Timeout()
	case *url.Error:
		// Only necessary for Go < 1.6.
		if err, ok := err.Err.(net.Error); ok {
			return err.Timeout()
		}
	}
	return false
}
