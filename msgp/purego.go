// +build purego appengine

package msgp

func UnsafeString(b []byte) string {
	return string(b)
}

func UnsafeBytes(s string) []byte {
	return []byte(s)
}
