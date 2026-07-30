package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerFont(pageObject *uint64) *FontHandle {
	ref := uuid.New()
	handle := &FontHandle{
		handle:    pageObject,
		nativeRef: references.FPDF_FONT(ref.String()),
	}

	p.fontRefs[handle.nativeRef] = handle

	return handle
}
