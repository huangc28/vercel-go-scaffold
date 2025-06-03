package middlewares

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

// responseLogger wraps http.ResponseWriter to capture status code and response size.
type responseLogger struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
	size   int
}

func newResponseLogger(w http.ResponseWriter) *responseLogger {
	return &responseLogger{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
	}
}

func (rl *responseLogger) WriteHeader(code int) {
	rl.status = code
	rl.ResponseWriter.WriteHeader(code)
}

func (rl *responseLogger) Write(b []byte) (int, error) {
	// Default status to 200 if WriteHeader wasn’t called explicitly.
	if rl.status == 0 {
		rl.status = http.StatusOK
	}

	// Write to actual ResponseWriter
	n, err := rl.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}
	rl.size += n

	// Copy into our buffer (but cap at, say 4KB)
	if rl.body.Len()+len(b) < 4096 {
		toWrite := b
		remaining := 4096 - rl.body.Len()
		if len(b) > remaining {
			toWrite = b[:remaining]
		}
		rl.body.Write(toWrite)
	}

	return n, err
}

// Logger is a chi middleware that logs request and response (including body).
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rl := newResponseLogger(w)

		// Call the wrapped handler
		next.ServeHTTP(rl, r)

		// After response is written, log everything
		log.Printf(
			"%s %s → %d (%d bytes) in %s\nResponse body (truncated to 4KB):\n%s\n",
			r.Method,
			r.URL.Path,
			rl.status,
			rl.size,
			time.Since(start),
			rl.body.String(),
		)
	})
}
