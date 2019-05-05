package common

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

func Alloc(size uintptr) *byte {
	return (*byte)(C.malloc(_Ctype_size_t(size)))
}

func Free(ptr *byte) {
	C.free(unsafe.Pointer(ptr))
}
