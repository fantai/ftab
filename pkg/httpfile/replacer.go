package httpfile

// Replacer transform key to value
type Replacer interface {
	// Get return the value by key, if not found key, should return "", false
	Get(key string) (string, bool)
}

// MapReplacer is a ValueExtractor based on map[string]string
type MapReplacer map[string]string

// ListReplacer is a ValueExtractor based on []ValueExtractor
type ListReplacer []Replacer

// Get is required by ValueExtractor interface
func (me MapReplacer) Get(key string) (string, bool) {
	val, ok := me[key]
	return val, ok
}

// Get is required by ValueExtractor interface
func (le ListReplacer) Get(key string) (string, bool) {
	for _, e := range le {
		if val, ok := e.Get(key); ok {
			return val, ok
		}
	}
	return "", false
}
