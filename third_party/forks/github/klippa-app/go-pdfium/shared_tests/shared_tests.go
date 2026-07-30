package shared_tests

import (
	"sync"

	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium"
)

var PdfiumInstance pdfium.Pdfium
var PdfiumPool pdfium.Pool
var TestDataPath string
var TestType string
var Locker = &sync.Mutex{} // A locker, sometimes we need to make sure things can't run concurrently.

func Import() {
	// We need this method to import the package into our different tests.
}
