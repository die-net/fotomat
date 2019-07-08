package thumbnail

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/die-net/fotomat/v2/format"
)

const (
	// DefaultAccept is the default Accept header sent on upstream requests.
	DefaultAccept = "image/jpeg,image/webp;q=0.8,image/png;q=0.4,image/svg+xml;q=0.3,image/gif;q=0.2,application/pdf;q=0.1"
	// DefaultServer is the default Server header sent with responses.
	DefaultServer = "Fotomat"
	// DefaultUserAgent is the default User-Agent header sent on upstream requests.
	DefaultUserAgent = "Fotomat (http://fotomat.org)"
)

// Proxy represents an HTTP proxy that can optionally run its contents
// through Thumbnail. Must be created with NewProxy.
type Proxy struct {
	Director  func(*http.Request) (Options, int)
	Client    *http.Client
	Accept    string
	Server    string
	UserAgent string
	pool      *Pool
	active    chan bool
}

// NewProxy creates a Proxy object, with a given Director, Pool, upper limit
// on images held in RAM, and http.Client.
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

// ServeHTTP serves an HTTP request for a given Proxy, using Director to
// parse the request, fetching an image, calling pool.Thumbnail on it, and
// returning the result.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, or *http.Request) {
	ctx := or.Context()

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

	if options.MaxQueueDuration <= 0 {
		options.MaxQueueDuration = time.Hour // "Forever" for an http request
	}

	// Wait for our turn to fetch and hold the original image.
	timeout := time.NewTimer(options.MaxQueueDuration)
	defer timeout.Stop()
	select {
	case <-ctx.Done():
		proxyError(w, ErrAborted, 0)
		return
	case <-timeout.C:
		proxyError(w, nil, http.StatusGatewayTimeout)
		return
	case <-p.active:
	}
	defer func() { p.active <- true }()

	orig, header, status, err := p.get(ctx, or.URL.String(), or.Header)
	if err != nil || (status != http.StatusOK && status != http.StatusNotModified) {
		proxyError(w, err, status)
		return
	}

	copyHeaders(header, w.Header(), []string{"Age", "Cache-Control", "Date", "Etag", "Expires", "Last-Modified"})
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	if status == http.StatusNotModified || isNotModified(or.Header, header) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	thumb, err := p.pool.Thumbnail(ctx, orig, options)
	if err != nil {
		proxyError(w, err, 0)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(thumb)))
	_, _ = w.Write(thumb)
}

func (p *Proxy) get(ctx context.Context, url string, header http.Header) ([]byte, http.Header, int, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	r = r.WithContext(ctx)

	// Pass some headers on to upstream.
	r.Header.Set("Accept", p.Accept)
	r.Header.Set("User-Agent", p.UserAgent)
	copyHeaders(header, r.Header, []string{"Cache-Control", "If-Modified-Since", "If-None-Match"})

	resp, err := p.Client.Do(r)
	if err != nil {
		return nil, nil, 0, err
	}

	orig, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	return orig, resp.Header, resp.StatusCode, err
}

// Close shuts down a Proxy.
func (p *Proxy) Close() {
	close(p.active)
	p.pool.Close()
	*p = Proxy{}
}

func copyHeaders(src, dest http.Header, keys []string) {
	for _, key := range keys {
		if value, ok := src[key]; ok {
			dest[key] = value
		}
	}
}

func isNotModified(req, resp http.Header) bool {
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
		http.StatusGatewayTimeout,
		http.StatusGone:
		err = nil
	case 0:
		switch err {
		case format.ErrUnknownFormat, ErrTooSmall:
			status = http.StatusUnsupportedMediaType
		case ErrTooBig:
			status = http.StatusRequestEntityTooLarge
		case ErrAborted, context.Canceled:
			status = 499 // Nginx error for "Client closed connection"
		default:
			if isTimeout(err) {
				err = nil
				status = http.StatusGatewayTimeout
			} else {
				status = http.StatusInternalServerError
			}
		}
	default:
		err = fmt.Errorf("proxy received %d %s", status, http.StatusText(status))
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
	if err, ok := err.(net.Error); ok {
		return err.Timeout()
	}
	return false
}
