package utils

import "unsafe"

func YoloString(b []byte) string {
	return *((*string)(unsafe.Pointer(&b)))
}
