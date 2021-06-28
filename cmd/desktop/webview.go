package main

/*
#include <stdlib.h>
#include <gtk/gtk.h>
#cgo pkg-config: gtk+-3.0
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/webview/webview"
)



// from https://github.com/gotk3/gotk3/blob/86f85cbecd0b990beab32a3471kb08ad3cdd8f93b/gtk/window.go#L531
func setWindowIcon(w webview.WebView, filename string) error {
	window := w.Window()
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))

	var err *C.GError = nil
	res := C.gtk_window_set_icon_from_file((*C.GtkWindow)(window), (*C.gchar)(cstr), &err)
	if res == 0 {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}