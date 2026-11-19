//go:build !windows

package pet

import (
	"errors"

	"github.com/ebitengine/purego"
)

// openVPXLibrary opens the system's libvpx. The engine plays the clips with core.Video on macOS and Linux, so only the decoder's tests come here, to run against a libvpx such as Homebrew's or a distribution's.
func openVPXLibrary() (uintptr, error) {
	var errs []error
	for _, name := range []string{"libvpx.dylib", "/opt/homebrew/lib/libvpx.dylib", "/usr/local/lib/libvpx.dylib", "libvpx.so", "libvpx.so.9", "libvpx.so.8", "libvpx.so.7"} {
		h, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err == nil {
			return h, nil
		}
		errs = append(errs, err)
	}
	return 0, errors.Join(errs...)
}
