package requests

import "github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

type GetActionInfo struct {
	Document references.FPDF_DOCUMENT
	Action   references.FPDF_ACTION
}
