package responses

import "github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"

type FPDFDoc_GetJavaScriptActionCount struct {
	JavaScriptActionCount int
}

type FPDFDoc_GetJavaScriptAction struct {
	Index            int
	JavaScriptAction references.FPDF_JAVASCRIPT_ACTION
}

type FPDFDoc_CloseJavaScriptAction struct{}

type FPDFJavaScriptAction_GetName struct {
	Name string
}

type FPDFJavaScriptAction_GetScript struct {
	Script string
}
