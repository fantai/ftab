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

	"github.com/fantai/ftab/pkg/mock"
	"github.com/valyala/fasthttp"
)

// Case is a http request case
type Case struct {
	Name           string             // name of case
	RespCode       int                // reponse code
	RequestSize    int                // request body length
	ResponseSize   int                // resoonse bytes length
	RespTime       time.Duration      // response time
	request        *fasthttp.Request  // the request object
	response       *fasthttp.Response // the responsee object
	parsedReqBody  interface{}        // parsed request body
	parsedRespBody interface{}        // parsed response body
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
	AutoClean bool              // automatic release resource, otherwise caller should do Release after use, default is true
}

// ###
var newCaseTag, _ = regexp.Compile(`^\s*###\s*$`)

// # @name=value
var nameTag, _ = regexp.Compile(`^\s*#\s+@name\s*=?\s*(\w+)\s*$`)

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

// DisaableAutoClean disable clear after Exectue
func DisaableAutoClean(f *HTTPFile) {
	f.AutoClean = false
}

// Opt is option when parse HTTPFile
type Opt func(f *HTTPFile)

// ParseReader parse httpfile from a reader
func ParseReader(r io.Reader, opts ...Opt) (*HTTPFile, error) {
	s := bufio.NewScanner(r)

	file := &HTTPFile{
		Variables: make(map[string]string),
		Cases:     make([]*Case, 0),
	}
	for _, opt := range opts {
		opt(file)
	}

	thisCase := &Case{
		request: fasthttp.AcquireRequest(),
	}
	stage := parseFileStage

	var groups [][]byte
	for s.Scan() {
		line := s.Bytes()

		if newCaseTag.Match(line) {
			file.Cases = append(file.Cases, thisCase)
			thisCase = &Case{
				request: fasthttp.AcquireRequest(),
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
			thisCase.request.Header.SetMethod(string(groups[1]))
			thisCase.request.SetRequestURI(string(groups[2]))

			continue
		}

		if stage < parseBodyStage {
			groups = headerDefineTag.FindSubmatch(line)
			if groups != nil {
				thisCase.request.Header.SetBytesKV(groups[1], groups[2])
				stage = parseHeaderStage
				continue
			}

			if stage == parseHeaderStage && (len(bytes.TrimSpace(line))) == 0 {
				stage = parseBodyStage
			}
		}

		if stage == parseBodyStage {
			thisCase.request.AppendBody(line)
		}
	}

	file.Cases = append(file.Cases, thisCase)

	return file, nil
}

// ParseFile parse httpfile from a file
func ParseFile(fileName string, opts ...Opt) (*HTTPFile, error) {
	fp, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("open %s failed: %w", fileName, err)
	}
	defer fp.Close()
	return ParseReader(fp, opts...)
}

// ParseBytes parse httpfile from content
func ParseBytes(content []byte, opts ...Opt) (*HTTPFile, error) {
	r := bytes.NewReader(content)
	return ParseReader(r, opts...)
}

// ReplaceVariable replace variable placeholder to value
func ReplaceVariable(text []byte, ve Replacer) []byte {
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
func ReplaceVariableString(text string, ve Replacer) string {
	return string(ReplaceVariable([]byte(text), ve))
}

// Duplicate this file, if mock is used, variable will be mocked
func (f *HTTPFile) Duplicate(useMock bool, expand bool) *HTTPFile {

	result := HTTPFile{
		Variables: make(map[string]string),
		Cases:     make([]*Case, len(f.Cases)),
		AutoClean: f.AutoClean,
	}
	for key, val := range f.Variables {
		if useMock {
			result.Variables[key] = mock.Value(key, val)
		} else {
			result.Variables[key] = val
		}
	}
	for i := 0; i < len(f.Cases); i++ {
		to := &Case{}
		from := f.Cases[i]

		to.Name = from.Name
		to.request = fasthttp.AcquireRequest()
		from.request.CopyTo(to.request)

		if expand {
			to.request.Header.SetMethodBytes(ReplaceVariable(to.request.Header.Method(), f))
			to.request.SetRequestURIBytes(ReplaceVariable(to.request.RequestURI(), f))
			to.request.SetBody(ReplaceVariable(to.request.Body(), f))
		}

		result.Cases[i] = to
	}
	return &result
}

// Release resource
func (f *HTTPFile) Release() {
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
}

// Execute the httfile
func (f *HTTPFile) Execute(client *fasthttp.Client, ve ...Replacer) error {
	lists := append(make(ListReplacer, 0), ve...)
	lists = append(lists, f)

	defer func() {
		if f.AutoClean {
			f.Release()
		}
	}()

	for _, to := range f.Cases {

		to.request.Header.SetMethodBytes(ReplaceVariable(to.request.Header.Method(), lists))
		to.request.SetRequestURIBytes(ReplaceVariable(to.request.RequestURI(), lists))
		to.request.SetBody(ReplaceVariable(to.request.Body(), lists))

		to.request.Header.VisitAll(func(key, val []byte) {
			to.request.Header.SetBytesKV(key, ReplaceVariable(val, lists))
		})

		to.response = fasthttp.AcquireResponse()

		t1 := time.Now()
		err := client.Do(to.request, to.response)
		if err != nil {
			return fmt.Errorf("request %s failed: %w", string(to.request.RequestURI()), err)
		}
		t2 := time.Now()

		// will keep valid after Execute
		to.RespCode = to.response.StatusCode()
		to.RespTime = t2.Sub(t1)
		to.RequestSize = len(to.request.Header.Header()) + len(to.request.Body()) + len(to.request.RequestURI())
		to.ResponseSize = len(to.response.Header.Header()) + len(to.response.Body())
		if to.RespCode != 200 {
			return fmt.Errorf("response is not 200")
		}
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
	case "$timestampms":
		return funTimestamapms(funcVar), true
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
				val := theCase.request.Header.Peek(name)
				if val != nil {
					return string(val), false
				}
			case "response":
				if theCase.response != nil {
					val := theCase.response.Header.Peek(name)
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
			json.Unmarshal(theCase.request.Body(), &theCase.parsedReqBody)
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
			return JSONPathGet(theCase.parsedRespBody, path), true
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
