package mock

import (
	"math/rand"
	"regexp"
	"time"
)

func newPatternMock(reText string, fn func() string) PatternMock {
	re, _ := regexp.Compile(reText)
	return PatternMock{re, fn}
}

func dateTimeMock(layout string) func() string {
	return func() string {
		now := time.Now()
		changed := rand.Int63n(60*int64(time.Hour)*24) - 30*int64(time.Hour)*24
		mocked := now.Add(time.Duration(changed))
		return mocked.Format(layout)
	}
}

func buildinPatterns() []PatternMock {
	return []PatternMock{
		newPatternMock(`^\d\d\d\d-\d\d-\d\d$`, dateTimeMock("2006-01-02")),
		newPatternMock(`^\d\d\d\d/\d\d/\d\d$`, dateTimeMock("2006/01/02")),
		newPatternMock(`^\d\d/\d\d/\d\d$`, dateTimeMock("06/01/02")),
		newPatternMock(`^\d\d:\d\d:\d\d$`, dateTimeMock("15:04:05")),
		newPatternMock(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\\.\d\d\d$`, dateTimeMock("2006-01-02T15:04:05.000")),
		newPatternMock(`^\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d$`, dateTimeMock("2006-01-02 15:04:05")),
	}
}
