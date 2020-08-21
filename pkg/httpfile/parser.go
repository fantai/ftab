package httpfile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aifantai/ftab/pkg/mock"
	"github.com/valyala/fasthttp"
)

// Case is a http request case
type Case struct {
	Name           string        // name of case
	Method         string        // the method of case
	URL            string        // the URL of case, URL maybe a variable
	RespCode       int           // reponse code
	RespBody       []byte        // resoonse bytes
	RespTime       time.Duration // response time
	ReqHeader      map[string]string
	ReqBody        *bytes.Buffer
	request        *fasthttp.Request
	response       *fasthttp.Response
	parsedReqBody  interface{}
	parsedRespBody interface{}
}

const (
	parseFileStage = iota
	parseHeaderStage
	parseBodyStage
)

// HTTPFile is a .http or .rest file parse result
type HTTPFile struct {
	Variables map[string]string // variable in this file
	Cases     []*Case           // all cases
}

// ###
var newCaseTag, _ = regexp.Compile(`^\s*###\s*$`)

// # @name=value
var nameTag, _ = regexp.Compile(`^\s*#\s+@name\s*=\s*(\w+)\s*$`)

// GET url
var firstLineTag, _ = regexp.Compile(`^\s*(GET|POST)\s+(.+?)\s*$`)

// @key=value
var variableDefineTag, _ = regexp.Compile(`^\s*@([[:graph:]]+)\s*=\s*(.+?)\s*$`)

// Content-Length: 123
var headerDefineTag, _ = regexp.Compile(`^\s*([[:graph:]]+)\s*:\s*(.+?)\s*$`)

// # this is comment
var commentTag, _ = regexp.Compile(`^\s*(#|//).*$`)

// {{abc}}
var variableRef, _ = regexp.Compile(`{{.+?}}`)

// ParseReader parse httpfile from a reader
func ParseReader(r io.Reader) (*HTTPFile, error) {
	s := bufio.NewScanner(r)

	file := &HTTPFile{
		Variables: make(map[string]string),
		Cases:     make([]*Case, 0),
	}

	thisCase := &Case{
		ReqHeader: make(map[string]string),
		ReqBody:   bytes.NewBuffer(nil),
	}
	stage := parseFileStage

	var groups [][]byte
	for s.Scan() {
		line := s.Bytes()

		if newCaseTag.Match(line) {
			file.Cases = append(file.Cases, thisCase)
			thisCase = &Case{
				ReqHeader: make(map[string]string),
				ReqBody:   bytes.NewBuffer(nil),
			}
			stage = parseFileStage
			continue
		}

		groups = variableDefineTag.FindSubmatch(line)
		if groups != nil {
			file.Variables[string(groups[1])] = string(groups[2])
			continue
		}

		groups = nameTag.FindSubmatch(line)
		if groups != nil {
			thisCase.Name = string(groups[1])
			continue
		}

		if commentTag.Match(line) {
			continue
		}

		groups = firstLineTag.FindSubmatch(line)
		if groups != nil {
			thisCase.Method = string(groups[1])
			thisCase.URL = string(groups[2])
			continue
		}

		if stage < parseBodyStage {
			groups = headerDefineTag.FindSubmatch(line)
			if groups != nil {
				thisCase.ReqHeader[string(groups[1])] = string(groups[2])
				stage = parseHeaderStage
				continue
			}

			if stage == parseHeaderStage && (len(bytes.TrimSpace(line))) == 0 {
				stage = parseBodyStage
			}
		}

		if stage == parseBodyStage {
			thisCase.ReqBody.Write(line)
		}
	}

	file.Cases = append(file.Cases, thisCase)

	return file, nil
}

// ParseFile parse httpfile from a file
func ParseFile(fileName string) (*HTTPFile, error) {
	fp, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("open %s failed: %w", fileName, err)
	}
	defer fp.Close()
	return ParseReader(fp)
}

// ParseBytes parse httpfile from content
func ParseBytes(content []byte) (*HTTPFile, error) {
	r := bytes.NewReader(content)
	return ParseReader(r)
}

// ValueExtractor transform key to value
type ValueExtractor interface {
	// Get return the value by key, if not found key, should return "", false
	Get(key string) (string, bool)
}

// MapValueExtractor is a ValueExtractor based on map[string]string
type MapValueExtractor map[string]string

// ListValueExtractor is a ValueExtractor based on []ValueExtractor
type ListValueExtractor []ValueExtractor

// Get is required by ValueExtractor interface
func (me MapValueExtractor) Get(key string) (string, bool) {
	val, ok := me[key]
	return val, ok
}

// Get is required by ValueExtractor interface
func (le ListValueExtractor) Get(key string) (string, bool) {
	for _, e := range le {
		if val, ok := e.Get(key); ok {
			return val, ok
		}
	}
	return "", false
}

// ReplaceVariable replace variable placeholder to value
func ReplaceVariable(text []byte, ve ValueExtractor) []byte {
	replaced := variableRef.ReplaceAllFunc([]byte(text), func(key []byte) []byte {
		rkey := bytes.TrimSpace(key[2 : len(key)-2])
		val, ok := ve.Get(string(rkey))
		if ok {
			return []byte(val)
		}
		return key
	})
	return replaced
}

// ReplaceVariableString replace variable placeholder to value string
func ReplaceVariableString(text string, ve ValueExtractor) string {
	return string(ReplaceVariable([]byte(text), ve))
}

// Duplicate this file, if mock is used, variable will be mocked
func (f *HTTPFile) Duplicate(useMock bool, expand bool) *HTTPFile {

	result := HTTPFile{
		Variables: make(map[string]string),
		Cases:     append([]*Case{}, f.Cases...),
	}
	for key, val := range f.Variables {
		if useMock {
			result.Variables[key] = mock.Value(key, val)
		} else {
			result.Variables[key] = val
		}
	}
	if expand {
		for _, c := range f.Cases {
			c.Method = ReplaceVariableString(c.Method, f)
			c.URL = ReplaceVariableString(c.URL, f)
			c.ReqBody = bytes.NewBuffer(ReplaceVariable(c.ReqBody.Bytes(), f))
		}
	}
	return &result
}

// Execute the httfile
func (f *HTTPFile) Execute(client *fasthttp.Client, ve ...ValueExtractor) error {
	lists := append(make(ListValueExtractor, 0), ve...)
	lists = append(lists, f)

	defer func() {
		for _, c := range f.Cases {
			if c.request != nil {
				fasthttp.ReleaseRequest(c.request)
			}
			if c.response != nil {
				fasthttp.ReleaseResponse(c.response)
			}
			c.request = nil
			c.response = nil
		}
	}()

	for _, c := range f.Cases {

		method := ReplaceVariableString(c.Method, lists)
		url := ReplaceVariableString(c.URL, lists)

		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()

		req.Header.SetMethod(method)
		req.SetRequestURI(url)

		for key, val := range c.ReqHeader {
			req.Header.Set(key, ReplaceVariableString(val, lists))
		}
		req.SetBody(ReplaceVariable(c.ReqBody.Bytes(), lists))

		t1 := time.Now()
		err := fasthttp.Do(req, resp)
		if err != nil {
			return fmt.Errorf("request %s failed: %w", url, err)
		}
		t2 := time.Now()

		c.RespCode = resp.StatusCode()
		c.RespBody = append([]byte{}, resp.Body()...)
		c.RespTime = t2.Sub(t1)

		// request , response maybe used in next case
		c.request = req
		c.response = resp
	}

	return nil
}

// Get is required by ValueExtractor interface
func (f *HTTPFile) Get(key string) (string, bool) {
	// variable with defined name
	val, ok := f.Variables[key]
	if ok {
		return val, ok
	}

	// variable buildin
	if key[0] == '$' {
		return f.getBuildinVariable(key)
	}

	// header variable
	if strings.Index(key, ".header.") > 0 {
		return f.getHeaderVariable(key)
	}

	if strings.Index(key, ".body.") > 0 {
		return f.getJSONPathVariable(key)
	}
	return val, ok
}

func (f *HTTPFile) getBuildinVariable(key string) (string, bool) {
	funcVar := strings.Split(key, " ")
	switch funcVar[0] {
	case "$timestamp":
		return funTimestamap(funcVar), true
	case "$randomInt":
		return funRandomInt(funcVar), true
	case "$datetime":
		return funDateTime(funcVar), true
	case "$localDatetime":
		return funLocalDateTime(funcVar), true
	}
	return "", false
}

func (f *HTTPFile) getHeaderVariable(key string) (string, bool) {
	args := strings.Split(key, ".")
	if len(args) == 4 {
		caseName, kind, name := args[0], args[1], args[3]
		theCase := f.findCaseByName(caseName)
		if theCase != nil {
			switch kind {
			case "request":
				val, ok := theCase.ReqHeader[name]
				return val, ok
			case "response":
				if theCase.response != nil {
					val := theCase.response.Header.Peek(key)
					if val != nil {
						return string(val), false
					}
				}
			}
		}
	}
	return "", false
}

func (f *HTTPFile) getJSONPathVariable(key string) (string, bool) {
	pos := strings.Index(key, ".body.")
	if pos < 0 {
		return "", false
	}
	pos += 6

	args := strings.Split(key[0:pos], ".")
	if len(args) < 3 {
		return "", false
	}

	caseName, kind := args[0], args[1]
	theCase := f.findCaseByName(caseName)
	if theCase == nil {
		return "", false
	}

	path := key[pos:]

	switch kind {
	case "request":
		if theCase.parsedReqBody == nil {
			json.Unmarshal(theCase.ReqBody.Bytes(), &theCase.parsedReqBody)
		}
		return JSONPathGet(theCase.parsedReqBody, path), true
	case "response":
		if theCase.response != nil {
			var body []byte
			switch string(theCase.response.Header.Peek("Content-Encoding")) {
			case "gzip":
				body, _ = theCase.response.BodyGunzip()
				break
			case "deflate":
				body, _ = theCase.response.BodyInflate()
			default:
				body = theCase.response.Body()
			}
			if theCase.parsedRespBody == nil {
				json.Unmarshal(body, &theCase.parsedRespBody)
			}
			return JSONPathGet(theCase.parsedReqBody, path), true
		}
	}
	return "", false
}

func (f *HTTPFile) findCaseByName(name string) *Case {
	for _, theCase := range f.Cases {
		if theCase.Name == name {
			return theCase
		}
	}
	return nil
}
