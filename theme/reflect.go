package theme

import "reflect"

// reflectTypeName returns the concrete type name of v, dereferencing one pointer level.
// Used by bus instrumentation; isolated here to keep provider.go imports tight.
func reflectTypeName(v any) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.String()
}
