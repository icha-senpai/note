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

package model

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const officeXMLPartMaxBytes = 20 << 20

type officeContentTypes struct {
	Overrides []officeContentOverride `xml:"Override"`
}

type officeContentOverride struct {
	ContentType string `xml:"ContentType,attr"`
	PartName    string `xml:"PartName,attr"`
}

func extractDocxText(absPath string) (string, error) {
	return extractOfficeOpenXMLText(absPath, map[string]officeTextPart{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml":        {kind: "header"},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml": {kind: "body"},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml":        {kind: "footer"},
	})
}

func extractPptxText(absPath string) (string, error) {
	return extractOfficeOpenXMLText(absPath, map[string]officeTextPart{
		"application/vnd.openxmlformats-officedocument.presentationml.slide+xml":  {kind: "body"},
		"application/vnd.openxmlformats-officedocument.drawingml.diagramData+xml": {kind: "body"},
	})
}

type officeTextPart struct {
	kind string
}

func extractOfficeOpenXMLText(absPath string, textParts map[string]officeTextPart) (string, error) {
	zr, err := zip.OpenReader(absPath)
	if err != nil {
		return "", fmt.Errorf("open Office archive: %w", err)
	}
	defer zr.Close()

	files := officeZipFiles(zr.File)
	contentTypesFile := files["[Content_Types].xml"]
	if contentTypesFile == nil {
		return "", fmt.Errorf("missing [Content_Types].xml")
	}

	contentTypes, err := readOfficeContentTypes(contentTypesFile)
	if err != nil {
		return "", err
	}

	var header, body, footer strings.Builder
	for _, override := range contentTypes.Overrides {
		part, ok := textParts[override.ContentType]
		if !ok {
			continue
		}

		file := files[strings.TrimPrefix(override.PartName, "/")]
		if file == nil {
			continue
		}

		text, err := readOfficeXMLText(file)
		if err != nil {
			return "", err
		}
		if "" == text {
			continue
		}

		switch part.kind {
		case "header":
			header.WriteString(text)
			header.WriteByte('\n')
		case "footer":
			footer.WriteString(text)
			footer.WriteByte('\n')
		default:
			body.WriteString(text)
			body.WriteByte('\n')
		}
	}

	return strings.TrimSpace(header.String() + "\n" + body.String() + "\n" + footer.String()), nil
}

func officeZipFiles(files []*zip.File) map[string]*zip.File {
	ret := make(map[string]*zip.File, len(files)*2)
	for _, file := range files {
		ret[file.Name] = file
		ret["/"+file.Name] = file
	}
	return ret
}

func readOfficeContentTypes(file *zip.File) (*officeContentTypes, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file.Name, err)
	}
	defer rc.Close()

	contentTypes := &officeContentTypes{}
	if err := xml.NewDecoder(io.LimitReader(rc, officeXMLPartMaxBytes)).Decode(contentTypes); err != nil {
		return nil, fmt.Errorf("parse %s: %w", file.Name, err)
	}
	return contentTypes, nil
}

func readOfficeXMLText(file *zip.File) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open %s: %w", file.Name, err)
	}
	defer rc.Close()

	text, err := officeXMLToText(rc)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", file.Name, err)
	}
	return text, nil
}

func officeXMLToText(r io.Reader) (string, error) {
	var ret bytes.Buffer
	dec := xml.NewDecoder(io.LimitReader(r, officeXMLPartMaxBytes))
	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		switch value := token.(type) {
		case xml.CharData:
			ret.Write(value)
		case xml.StartElement:
			switch value.Name.Local {
			case "br", "p", "tab":
				ret.WriteByte('\n')
			case "instrText", "script":
				if err := skipOfficeXMLNode(dec); err != nil {
					return "", err
				}
			}
		}
	}
	return ret.String(), nil
}

func skipOfficeXMLNode(dec *xml.Decoder) error {
	depth := 1
	for depth > 0 {
		token, err := dec.Token()
		if err != nil {
			return err
		}

		switch token.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}
