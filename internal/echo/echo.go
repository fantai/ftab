package echo

import (
	"bytes"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func echoHandler(ctx *fasthttp.RequestCtx) {

	if viper.GetBool("echo.verbose") {
		logger := zap.L()
		logger.Info("request", zap.Strings("content", strings.Split(ctx.Request.String(), "\r")))
	}

	ctx.Request.Header.VisitAll(func(key, val []byte) {
		if bytes.HasPrefix(key, []byte("echo-")) {
			ctx.Response.Header.AddBytesKV(key, val)
		}
	})
	ctx.Response.Header.Set("content-type", "application/json; charset=utf-8")
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
