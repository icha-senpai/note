package requests

import "github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

type GetBookmarks struct {
	Document references.FPDF_DOCUMENT
}
