package echo

import (
	"bytes"
	"sync"

	"github.com/valyala/fasthttp"
)

func echoHandler(ctx *fasthttp.RequestCtx) {
	ctx.Request.Header.VisitAll(func(key, val []byte) {
		if bytes.HasPrefix(key, []byte("Echo-")) {
			ctx.Response.Header.AddBytesKV(key, val)
		}
	})
	ctx.Response.SetBody(ctx.Request.Body())
	ctx.Response.SetStatusCode(200)
}

// Start run a echo server
func Start(addr string, wg *sync.WaitGroup) error {

	s := &fasthttp.Server{
		Handler: echoHandler,
		Name:    "echo server",
	}
	var err error

	go func() {
		err = s.ListenAndServe(addr)
		if err != nil {
			wg.Done()
		}
	}()
	wg.Wait()
	if err != nil {
		return err
	}
	return s.Shutdown()
}
