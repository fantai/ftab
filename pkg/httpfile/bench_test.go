package httpfile

import (
	"fmt"
	"os"
	"testing"
)

func TestBench(t *testing.T) {
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
		t.Error(err)
	}

	stats, timeUsed := Bench(file, 100, 2000)

	//t.Log(stats)

	report := ReportStat(stats, timeUsed)

	/*
		text, err := json.MarshalIndent(&report, "", "  ")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(text))
	*/

	HumanOutput(&report, os.Stdout)
}
