package httpfile

import (
	"bytes"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/ratelimit"
)

var client = &fasthttp.Client{
	MaxConnsPerHost: 2048,
	ReadTimeout:     time.Second,
	WriteTimeout:    time.Second,
}

var rateLimiter ratelimit.Limiter

func executeN(file *HTTPFile, n int, done chan bool, stats chan Stat) {

	for i := 0; i < n; i++ {
		if rateLimiter != nil {
			rateLimiter.Take()
		}

		w := file.Duplicate(true, true)
		err := w.Execute(client)

		var stat Stat
		if err != nil {
			stat.Failed = 1
		} else {
			stat.Successed = 1
		}
		stat.Requests = stat.Successed + stat.Failed
		for _, c := range w.Cases {
			stat.BytesSend = stat.BytesSend + c.RequestSize
			stat.BytesReceived = stat.BytesReceived + c.ResponseSize
			stat.TimeConsuming = stat.TimeConsuming + c.RespTime.Seconds()
		}
		w.Release()
		stats <- stat
	}
	done <- true
}

// Bench the httpfile
func Bench(file *HTTPFile, connections, requests, rateLimit int) ([]Stat, float64) {

	stats := make(chan Stat, 1024)
	done := make(chan bool, connections)
	doneCounter := 0

	requestsPerConnection := requests / connections

	if rateLimit > 0 {
		rateLimiter = ratelimit.New(rateLimit)
	} else {
		rateLimiter = nil
	}

	for c := 0; c < connections; c++ {
		go executeN(file, requestsPerConnection, done, stats)
	}
	results := make([]Stat, 0)
	t1 := time.Now()
	for {
		select {
		case s := <-stats:
			results = append(results, s)
		case <-done:
			doneCounter = doneCounter + 1
			if doneCounter == connections {
				t2 := time.Now()
				return results, t2.Sub(t1).Seconds()
			}
		}
	}
}

// Execute the file once
func Execute(file *HTTPFile) string {
	client := &fasthttp.Client{}
	w := file.Duplicate(true, true)
	err := w.Execute(client)
	if err != nil {
		return err.Error()
	}
	buff := bytes.NewBuffer(nil)
	for _, c := range w.Cases {
		buff.Write(c.request.Header.Header())
		buff.WriteString("\r\n")
		buff.Write(c.request.Body())

		buff.WriteString("\r\n")
		buff.WriteString("\r\n")

		buff.Write(c.response.Header.Header())
		buff.WriteString("\r\n")
		buff.Write(c.response.Body())
	}
	return buff.String()
}
