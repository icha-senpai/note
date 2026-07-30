package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"

	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerLink(dest *uint64) *LinkHandle {
	ref := uuid.New()
	handle := &LinkHandle{
		handle:    dest,
		nativeRef: references.FPDF_LINK(ref.String()),
	}

	p.linkRefs[handle.nativeRef] = handle

	return handle
}
