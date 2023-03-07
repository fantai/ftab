package httpfile

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"
	"k8s.io/client-go/util/jsonpath"
)

func objFromString(text []byte) interface{} {
	var parsed interface{}
	err := json.Unmarshal([]byte(text), &parsed)
	if err != nil {
		zap.L().Error("jsonpath parse failed", zap.Error(err))
		return nil
	}
	return parsed
}

// JSONPathGet get a pathed value from data
func JSONPathGet(data interface{}, path string) string {
	if data == nil {
		return ""
	}
	var err error

	switch text := data.(type) {
	case string:
		data = objFromString([]byte(text))
	case []byte:
		data = objFromString([]byte(text))
	}
	if data == nil {
		return ""
	}

	p := jsonpath.New("")
	p.AllowMissingKeys(true)
	err = p.Parse("{" + path + "}")
	if err != nil {
		zap.L().Error("jsonpath parse failed", zap.String("path", path), zap.Error(err))
		return ""
	}
	buf := new(bytes.Buffer)
	err = p.Execute(buf, data)
	if err != nil {
		zap.L().Error("jsonpath exectue failed", zap.Error(err))
		return ""
	}
	return buf.String()
}
