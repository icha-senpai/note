package responses

import (
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/enums"
	"github.com/icha-senpai/note/third_party/forks/github/klippa-app/go-pdfium/references"
)

type ActionInfo struct {
	Reference references.FPDF_ACTION
	Type      enums.FPDF_ACTION_ACTION
	DestInfo  *DestInfo // Is set when the action is GOTO. When the action is REMOTEGOTO, we will not fetch the destination.
	FilePath  *string   // When action is LAUNCH or REMOTEGOTO.
	URIPath   *string   // When action is URI.
}

type GetActionInfo struct {
	ActionInfo ActionInfo
}
