# libvpx for Windows

The Windows build decodes the WebM clips with libvpx, which NSGo's Windows player cannot do with their alpha. These files come unchanged from MSYS2's mingw64 packages, fetched by `tools/fetch-libvpx.sh`:

| File | Package | License |
| --- | --- | --- |
| `libvpx-1.dll` | [mingw-w64-x86_64-libvpx-1.16.0-1-any.pkg.tar.zst](https://mirror.msys2.org/mingw/mingw64/mingw-w64-x86_64-libvpx-1.16.0-1-any.pkg.tar.zst), libvpx 1.16.0 by the WebM Project | BSD-3-Clause, with Google's patent grant: [libvpx-LICENSE](libvpx-LICENSE), [libvpx-PATENTS](libvpx-PATENTS) |
| `libwinpthread-1.dll` | [mingw-w64-x86_64-libwinpthread-14.0.0.r92.g818fa6510-1-any.pkg.tar.zst](https://mirror.msys2.org/mingw/mingw64/mingw-w64-x86_64-libwinpthread-14.0.0.r92.g818fa6510-1-any.pkg.tar.zst), winpthreads by the mingw-w64 project, which libvpx-1.dll links | MIT and others: [libwinpthread-COPYING](libwinpthread-COPYING) |

The libvpx license and patent grant are those of its [v1.16.0](https://github.com/webmproject/libvpx/tree/v1.16.0) tag, which the package does not carry.
