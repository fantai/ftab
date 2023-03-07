package httpfile

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

const echoServer = "http://127.0.0.1:6601/"

func TestMain(m *testing.M) {
	//wg := &sync.WaitGroup{}
	//wg.Add(1)
	//go echo.Start(host, wg)
	//time.Sleep(time.Second)
	m.Run()
}
func TestParse(t *testing.T) {
	content := `
	@server = http://www.baidu.com

	POST {{server}}
	Content-Type: application/json
	
	{
		"a": "b"
	}
	###
	
	# @name=hello
	
	GET {{server}}
	`

	file, err := ParseBytes([]byte(content), DisaableAutoClean)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(file.Variables), "found 1 variable")
	assert.Equal(t, "http://www.baidu.com", file.Variables["server"], "server is http://www.baidu.com")
	server, _ := file.Get("server")
	assert.Equal(t, "http://www.baidu.com", server)

	assert.Equal(t, 2, len(file.Cases))
	assert.Equal(t, "POST", string(file.Cases[0].request.Header.Method()))
	assert.Equal(t, "application/json", string(file.Cases[0].request.Header.ContentType()))
	assert.Equal(t, "{{server}}", string(file.Cases[0].request.Header.RequestURI()))
	assert.Equal(t, "http://www.baidu.com", string(ReplaceVariable(file.Cases[0].request.Header.RequestURI(), file)))

	assert.JSONEq(t, `{"a":"b"}`, string(file.Cases[0].request.Body()))

	assert.Equal(t, "GET", string(file.Cases[1].request.Header.Method()))
	assert.Equal(t, "hello", file.Cases[1].Name)

	file.Release()
}

func TestExecute(t *testing.T) {
	content := fmt.Sprintf(`
	@server = %s

	# @name=case1
	POST {{server}}
	Content-Type: application/json
	
	{
		"a": "b"
	}
	###
	
	# @name=hello
	
	POST {{server}}
	Content-Type: application/json

	{
		"a1": "{{case1.request.body.$.a}}",
		"a2": "{{case1.response.body.$.a}}"
	}

	`, echoServer)

	file, err := ParseBytes([]byte(content), DisaableAutoClean)
	if err != nil {
		t.Error(err)
	}

	err = file.Execute(&fasthttp.Client{})
	if err != nil {
		t.Error(err)
	}

	respDom := make(map[string]interface{})
	err = json.Unmarshal(file.Cases[1].response.Body(), &respDom)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 2, len(respDom))
	assert.Equal(t, "b", respDom["a1"])
	assert.Equal(t, "b", respDom["a2"])

	file.Release()
}

func BenchmarkExecute(b *testing.B) {
	content := fmt.Sprintf(`
	@server = %s

	# @name=case1
	POST {{server}}
	Content-Type: application/json
	
	{
		"a": "b"
	}
	###
	
	# @name=hello
	
	POST {{server}}
	Content-Type: application/json

	{
		"a1": "{{case1.request.body.$.a}}",
		"a2": "{{case1.response.body.$.a}}"
	}

	`, echoServer)

	file, err := ParseBytes([]byte(content))
	if err != nil {
		b.Error(err)
	}

	client := &fasthttp.Client{}

	for i := 0; i < b.N; i++ {
		err = file.Duplicate(true, true).Execute(client)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkRaw(b *testing.B) {

	client := &fasthttp.Client{}

	cases := []string{
		`
		{
			"a": "b"
		}
		`,
		`
		{
			"a1": "{{case1.request.body.$.a}}",
			"a2": "{{case1.response.body.$.a}}"
		}
		`,
	}

	for i := 0; i < b.N; i++ {
		for _, body := range cases {
			req := fasthttp.AcquireRequest()
			resp := fasthttp.AcquireResponse()

			req.SetBodyString(body)
			req.SetRequestURI(echoServer)
			req.Header.SetMethod("POST")

			client.Do(req, resp)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}
	}
}
