package httpfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeLayout(t *testing.T) {
	assert.Equal(t, "2006-01-02 03:04:05", layoutFromTimeFormat("YYYY-MM-DD HH:mm:ss"))
	assert.Equal(t, "2006-01-02T03:04:05.000", layoutFromTimeFormat("YYYY-MM-DDTHH:mm:ss.SSS"))
	assert.Equal(t, "2006-01-02T03:04:05.000 -0700", layoutFromTimeFormat("YYYY-MM-DDTHH:mm:ss.SSS ZZ"))
}

func TestJSONPath(t *testing.T) {
	text := `
	{
		"a": "b"
	}
	`

	assert.Equal(t, "b", JSONPathGet(text, "$.a"))
}
