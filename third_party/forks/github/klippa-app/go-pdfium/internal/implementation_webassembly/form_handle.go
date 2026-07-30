package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerFormHandle(formHandle *uint64, formInfo *uint64) *FormHandleHandle {
	ref := uuid.New()
	handle := &FormHandleHandle{
		handle:           formHandle,
		nativeRef:        references.FPDF_FORMHANDLE(ref.String()),
		formInfo:         formInfo,
		pagePointers:     map[uint64]references.FPDF_PAGE{},
		documentPointers: map[uint64]references.FPDF_DOCUMENT{},
	}

	p.formHandleRefs[handle.nativeRef] = handle

	return handle
}
