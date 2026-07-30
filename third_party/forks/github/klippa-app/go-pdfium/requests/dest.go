package requests

import "github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

type GetDestInfo struct {
	Document references.FPDF_DOCUMENT
	Dest     references.FPDF_DEST
}
