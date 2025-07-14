package main

/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: wrapper.o -lstdc++ -ltag
#include "wrapper.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <audiofile>")
		return
	}

	cPath := C.CString(os.Args[1])
	defer C.free(unsafe.Pointer(cPath))

	meta := C.read_metadata(cPath)
	if meta == nil {
		fmt.Println("Failed to read metadata")
		return
	}
	defer C.free_metadata(meta)

	fmt.Println("Title: ", C.GoString(meta.title))
	fmt.Println("Artist:", C.GoString(meta.artist))
	fmt.Println("Album: ", C.GoString(meta.album))
	fmt.Println("Year:  ", int(meta.year))
}

