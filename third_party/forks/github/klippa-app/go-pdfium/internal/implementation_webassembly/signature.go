package implementation_webassembly

import (
	"github.com/icha-senpai/note/third_party/forks/github/google/uuid"
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

func (p *PdfiumImplementation) registerSignature(signature *uint64, documentHandle *DocumentHandle) *SignatureHandle {
	ref := uuid.New()
	handle := &SignatureHandle{
		handle:      signature,
		nativeRef:   references.FPDF_SIGNATURE(ref.String()),
		documentRef: documentHandle.nativeRef,
	}

	documentHandle.signatureRefs[handle.nativeRef] = handle
	p.signatureRefs[handle.nativeRef] = handle

	return handle
}
