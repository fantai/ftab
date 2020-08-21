package mock

import (
	"math/rand"
	"testing"
	"time"
)

func TestCN(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	m := Load("cn")

	t.Log(m.EMail(), m.IDCard(), m.Name(), m.Mobile())
}

func TestMockValue(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	t.Log(Value("time", "2020-08-23"))
	t.Log(Value("time", "2020-08-23 15:23:42"))

}
