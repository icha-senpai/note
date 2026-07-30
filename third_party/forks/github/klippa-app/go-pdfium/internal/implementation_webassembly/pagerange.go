package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
)

func (p *PdfiumImplementation) registerPageRange(pageRange *uint64, documentHandle *DocumentHandle) *PageRangeHandle {
	ref := uuid.New()
	handle := &PageRangeHandle{
		handle:      pageRange,
		nativeRef:   references.FPDF_PAGERANGE(ref.String()),
		documentRef: documentHandle.nativeRef,
	}

	p.pageRangeRefs[handle.nativeRef] = handle
	documentHandle.pageRangeRefs[handle.nativeRef] = handle

	return handle
}
