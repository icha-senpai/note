package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"

	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerGlyphPath(glyphPath *uint64) *GlyphPathHandle {
	ref := uuid.New()
	handle := &GlyphPathHandle{
		handle:    glyphPath,
		nativeRef: references.FPDF_GLYPHPATH(ref.String()),
	}

	p.glyphPathRefs[handle.nativeRef] = handle

	return handle
}
