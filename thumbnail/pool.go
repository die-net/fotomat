package thumbnail

import (
	"errors"
	"github.com/die-net/fotomat/vips"
	"runtime"
	"sync"
)

var (
	ErrAborted = errors.New("Thumbnail request aborted")
)

// Pool represents a Thumbnail worker pool. VIPS keeps thread-local caches,
// which we retain control over through a combination of a pool of worker
// goroutines and using runtime.LockOSThread() within those workers.
type Pool struct {
	RequestCh chan *Request
	wg        sync.WaitGroup
}

// NewPool creates a Thumbnail worker pool with a given number of worker
// threads and queue length.
func NewPool(workers, queueLen int) *Pool {
	p := &Pool{RequestCh: make(chan *Request, queueLen)}

	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	for i := 0; i < workers; i++ {
		go p.worker()
	}

	return p
}

// Request to be sent to Pool.RequestCh to queue a Thumbnail operation.
type Request struct {
	Blob       []byte
	Options    Options
	Aborted    <-chan bool
	ResponseCh chan<- *Response
}

// Response sent to Request.ResponseCh when the Thumbnail operation is done.
type Response struct {
	Blob  []byte
	Error error
}

// Thumbnail is a blocking wrapper that executes thumbnail.Thumbnail
// requests in a pool of worker threads.  Work is skipped if aborted is
// closed while the request is queued.
func (p *Pool) Thumbnail(blob []byte, options Options, aborted <-chan bool) ([]byte, error) {
	rc := make(chan *Response)

	r := &Request{Blob: blob, Options: options, Aborted: aborted, ResponseCh: rc}
	p.RequestCh <- r

	s := <-rc
	close(rc)

	return s.Blob, s.Error
}

func (p *Pool) worker() {
	p.wg.Add(1)

	runtime.LockOSThread()

	for {
		q := <-p.RequestCh
		if q == nil {
			break
		}

		s := &Response{}
		if hasAborted(q.Aborted) {
			s.Error = ErrAborted
		} else {
			s.Blob, s.Error = Thumbnail(q.Blob, q.Options)
		}

		q.ResponseCh <- s
	}

	vips.ThreadShutdown()
	runtime.UnlockOSThread()

	p.wg.Done()
}

// Close shuts down the worker pool and waits for remaining work to be done.
func (p *Pool) Close() {
	close(p.RequestCh)
	p.wg.Wait()
}

func hasAborted(aborted <-chan bool) bool {
	select {
	case <-aborted:
		return true
	default:
		return false
	}
}
