package responses

import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

type DestInfo struct {
	Reference references.FPDF_DEST
	PageIndex int
}

type GetDestInfo struct {
	DestInfo DestInfo
}
