package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
)

func (p *PdfiumImplementation) registerPageObjectMark(pageObjectMark *uint64) *PageObjectMarkHandle {
	ref := uuid.New()
	handle := &PageObjectMarkHandle{
		handle:    pageObjectMark,
		nativeRef: references.FPDF_PAGEOBJECTMARK(ref.String()),
	}

	p.pageObjectMarkRefs[handle.nativeRef] = handle

	return handle
}
