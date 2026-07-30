package implementation_cgo

// #cgo pkg-config: pdfium
// #include "fpdf_transformpage.h"
import "C"
import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
)

func (p *PdfiumImplementation) registerPathSegment(pathSegment C.FPDF_PATHSEGMENT) *PathSegmentHandle {
	ref := uuid.New()
	handle := &PathSegmentHandle{
		handle:    pathSegment,
		nativeRef: references.FPDF_PATHSEGMENT(ref.String()),
	}

	p.pathSegmentRefs[handle.nativeRef] = handle

	return handle
}
