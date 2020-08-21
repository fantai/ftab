package mock

import (
	"regexp"
	"sort"
	"strings"

	"github.com/aifantai/ftab/pkg/mock/cn"
	"github.com/spf13/viper"
)

// Mocker is a data generator
type Mocker interface {
	IDCard() string
	EMail() string
	Mobile() string
	Name() string
}

// Load mocker for given name
// each country maybe have it's own mocker rules
func Load(name string) Mocker {
	switch name {
	case "cn":
		return &cn.Mocker{}
	default:
		return &cn.Mocker{}
	}
}

var dontMock = []string{
	"host",
	"port",
	"server",
}

// PatternMock is a generator for a pattern
type PatternMock struct {
	Pattern *regexp.Regexp
	Mock    func() string
}

var patterns []PatternMock

func init() {
	patterns = buildinPatterns()
}

// Value return a mock data for given name
func Value(name, originValue string) string {

	nameLower := strings.ToLower(name)

	if sort.SearchStrings(dontMock, nameLower) != len(dontMock) {
		return originValue
	}

	mocker := Load(viper.GetString("mocker"))
	switch nameLower {
	case "idcard":
		return mocker.IDCard()
	case "email":
		return mocker.EMail()
	case "name":
		return mocker.Name()
	case "mobile":
		return mocker.Mobile()
	}

	originValueBytes := []byte(originValue)
	for _, pm := range patterns {
		if pm.Pattern.Match(originValueBytes) {
			return pm.Mock()
		}
	}

	return originValue
}
