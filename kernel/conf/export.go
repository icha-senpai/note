// Scribli - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package conf

type Export struct {
	ParagraphBeginningSpace bool `json:"paragraphBeginningSpace"`
	AddTitle                bool `json:"addTitle"`

	BlockRefMode          int    `json:"blockRefMode"`
	BlockEmbedMode        int    `json:"blockEmbedMode"`
	BlockRefTextLeft      string `json:"blockRefTextLeft"`
	BlockRefTextRight     string `json:"blockRefTextRight"`
	TagOpenMarker         string `json:"tagOpenMarker"`
	TagCloseMarker        string `json:"tagCloseMarker"`
	FileAnnotationRefMode int    `json:"fileAnnotationRefMode"`
	PandocBin             string `json:"pandocBin"`
	PandocParams          string `json:"pandocParams"`
	DocxTemplate          string `json:"docxTemplate"`
	RemoveAssetsID        bool   `json:"removeAssetsID"`
	MarkdownYFM           bool   `json:"markdownYFM"`
	InlineMemo            bool   `json:"inlineMemo"`
	IncludeSubDocs        bool   `json:"includeSubDocs"`
	IncludeRelatedDocs    bool   `json:"includeRelatedDocs"`
	PDFFooter             string `json:"pdfFooter"`
	PDFWatermarkStr       string `json:"pdfWatermarkStr"`
	PDFWatermarkDesc      string `json:"pdfWatermarkDesc"`
	ImageWatermarkStr     string `json:"imageWatermarkStr"`
	ImageWatermarkDesc    string `json:"imageWatermarkDesc"`
}

func NewExport() *Export {
	return &Export{
		ParagraphBeginningSpace: false,
		AddTitle:                true,
		BlockRefMode:            4,
		BlockEmbedMode:          1,
		BlockRefTextLeft:        "",
		BlockRefTextRight:       "",
		TagOpenMarker:           "#",
		TagCloseMarker:          "#",
		FileAnnotationRefMode:   0,
		PandocBin:               "",
		RemoveAssetsID:          false,
		MarkdownYFM:             false,
		InlineMemo:              false,
		IncludeSubDocs:          true,
		IncludeRelatedDocs:      false,
		PDFFooter:               "%page / %pages",
	}
}
