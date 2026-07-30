package implementation_cgo

// #cgo pkg-config: pdfium
// #include "fpdf_transformpage.h"
import "C"
import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
)

func (p *PdfiumImplementation) registerClipPath(clipPath C.FPDF_CLIPPATH) *ClipPathHandle {
	ref := uuid.New()
	handle := &ClipPathHandle{
		handle:    clipPath,
		nativeRef: references.FPDF_CLIPPATH(ref.String()),
	}

	p.clipPathRefs[handle.nativeRef] = handle

	return handle
}
