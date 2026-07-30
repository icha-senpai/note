package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
)

func (p *PdfiumImplementation) registerClipPath(clipPath *uint64) *ClipPathHandle {
	ref := uuid.New()
	handle := &ClipPathHandle{
		handle:    clipPath,
		nativeRef: references.FPDF_CLIPPATH(ref.String()),
	}

	p.clipPathRefs[handle.nativeRef] = handle

	return handle
}
