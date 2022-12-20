package strings

import (
	"reflect"
	"unsafe"
)

// copy from prometheus source code

// NoAllocString convert []byte to string
func NoAllocString(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

// NoAllocBytes convert string to []byte
func NoAllocBytes(buf string) []byte {
	x := (*reflect.StringHeader)(unsafe.Pointer(&buf))
	h := reflect.SliceHeader{Data: x.Data, Len: x.Len, Cap: x.Len}
	return *(*[]byte)(unsafe.Pointer(&h))
}
