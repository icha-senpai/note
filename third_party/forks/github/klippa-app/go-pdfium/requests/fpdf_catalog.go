package requests

import "github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

type FPDFCatalog_IsTagged struct {
	Document references.FPDF_DOCUMENT
}

type FPDFCatalog_SetLanguage struct {
	Document references.FPDF_DOCUMENT
	Language string
}
