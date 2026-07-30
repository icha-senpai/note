package implementation_cgo

// #cgo pkg-config: pdfium
// #include "fpdf_formfill.h"
import "C"
import (
	"unsafe"

	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerFormHandle(formHandle C.FPDF_FORMHANDLE, formInfo unsafe.Pointer) *FormHandleHandle {
	ref := uuid.New()
	handle := &FormHandleHandle{
		handle:           formHandle,
		nativeRef:        references.FPDF_FORMHANDLE(ref.String()),
		formInfo:         formInfo,
		pagePointers:     map[unsafe.Pointer]references.FPDF_PAGE{},
		documentPointers: map[unsafe.Pointer]references.FPDF_DOCUMENT{},
	}

	p.formHandleRefs[handle.nativeRef] = handle

	return handle
}
