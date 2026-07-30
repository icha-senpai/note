package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"

	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerPageObject(pageObject *uint64) *PageObjectHandle {
	ref := uuid.New()
	handle := &PageObjectHandle{
		handle:    pageObject,
		nativeRef: references.FPDF_PAGEOBJECT(ref.String()),
	}

	p.pageObjectRefs[handle.nativeRef] = handle

	return handle
}
