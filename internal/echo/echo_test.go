package echo

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestEchoServer(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	host := "127.0.0.1:6601"

	go func() {
		err := Start(host, wg)
		if err != nil {
			fmt.Println("start server failed, %w", err)
		}
	}()
	// wait server start ready
	time.Sleep(time.Second)

	content := []byte(`{"a":","b"}`)

	req := &fasthttp.Request{}
	resp := &fasthttp.Response{}

	req.Header.SetMethod("POST")
	req.Header.Add("Echo-Key1", "Key1")
	req.AppendBody(content)
	req.SetRequestURI(fmt.Sprintf("http://%s/", host))

	err := fasthttp.Do(req, resp)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, content, resp.Body())
	assert.Equal(t, []byte("Key1"), resp.Header.Peek("Echo-Key1"))

	wg.Done()

}
