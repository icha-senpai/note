package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerSearch(search *uint64, documentHandle *DocumentHandle) *SearchHandle {
	ref := uuid.New()
	handle := &SearchHandle{
		handle:      search,
		nativeRef:   references.FPDF_SCHHANDLE(ref.String()),
		documentRef: documentHandle.nativeRef,
	}

	documentHandle.searchRefs[handle.nativeRef] = handle
	p.searchRefs[handle.nativeRef] = handle

	return handle
}
