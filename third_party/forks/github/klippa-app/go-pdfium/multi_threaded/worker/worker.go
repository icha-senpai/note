package worker

import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium"
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/internal/implementation_cgo"
)

func StartWorker(config *pdfium.LibraryConfig) {
	implementation_cgo.StartPlugin(config)
}
