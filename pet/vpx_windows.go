package pet

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

// loadLibraryExW is kernel32's LoadLibraryExW, which syscall.LoadDLL does not offer with flags.
var loadLibraryExW = syscall.NewLazyDLL("kernel32.dll").NewProc("LoadLibraryExW")

// The flags of LoadLibraryExW that look for a library's own dependencies in its folder and then in System32 only.
const (
	loadLibrarySearchDLLLoadDir = 0x100
	loadLibrarySearchSystem32   = 0x800
)

// openVPXLibrary opens the libvpx the host ships in its resources/windows, as nsgo-pet does, which NSGo's Windows packages put beside the executable (resourceRoot). Its winpthreads library is beside it, and the flags find it there rather than one of the same name elsewhere on PATH.
func openVPXLibrary() (uintptr, error) {
	dir, err := filepath.Abs(filepath.Join(resourceRoot, "windows"))
	if err != nil {
		return 0, err
	}
	path, err := syscall.UTF16PtrFromString(filepath.Join(dir, "libvpx-1.dll"))
	if err != nil {
		return 0, err
	}
	h, _, err := loadLibraryExW.Call(uintptr(unsafe.Pointer(path)), 0, loadLibrarySearchDLLLoadDir|loadLibrarySearchSystem32)
	if h == 0 {
		return 0, err
	}
	return h, nil
}
