// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
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
	"path"
	"strconv"
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

type xlsxWorkbook struct {
	Sheets []xlsxSheet `xml:"sheets>sheet"`
}

type xlsxSheet struct {
	Name string `xml:"name,attr"`
	RID  string `xml:"id,attr"`
}

type xlsxRelationships struct {
	Relationships []xlsxRelationship `xml:"Relationship"`
}

type xlsxRelationship struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
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

func extractXlsxText(absPath string) (string, error) {
	zr, err := zip.OpenReader(absPath)
	if err != nil {
		return "", fmt.Errorf("open Office archive: %w", err)
	}
	defer zr.Close()

	files := officeZipFiles(zr.File)
	sharedStrings, err := readXlsxSharedStrings(files["xl/sharedStrings.xml"])
	if err != nil {
		return "", err
	}

	sheetPaths, err := readXlsxSheetPaths(files)
	if err != nil {
		return "", err
	}
	if len(sheetPaths) == 0 {
		for name := range files {
			if strings.HasPrefix(name, "xl/worksheets/sheet") && strings.HasSuffix(name, ".xml") {
				sheetPaths = append(sheetPaths, name)
			}
		}
	}

	var ret strings.Builder
	for _, sheetPath := range sheetPaths {
		file := files[sheetPath]
		if file == nil {
			continue
		}
		text, err := readXlsxSheetText(file, sharedStrings)
		if err != nil {
			return "", err
		}
		if text != "" {
			ret.WriteString(text)
			ret.WriteByte(' ')
		}
	}
	return strings.TrimSpace(ret.String()), nil
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

func readXlsxSheetPaths(files map[string]*zip.File) ([]string, error) {
	workbookFile := files["xl/workbook.xml"]
	relationshipsFile := files["xl/_rels/workbook.xml.rels"]
	if workbookFile == nil || relationshipsFile == nil {
		return nil, nil
	}

	workbook := &xlsxWorkbook{}
	if err := decodeZipXML(workbookFile, workbook); err != nil {
		return nil, err
	}

	relationships := &xlsxRelationships{}
	if err := decodeZipXML(relationshipsFile, relationships); err != nil {
		return nil, err
	}

	targets := map[string]string{}
	for _, rel := range relationships.Relationships {
		target := strings.ReplaceAll(rel.Target, "\\", "/")
		if !strings.HasPrefix(target, "xl/") {
			target = path.Clean("xl/" + target)
		}
		targets[rel.ID] = target
	}

	var ret []string
	for _, sheet := range workbook.Sheets {
		if target := targets[sheet.RID]; target != "" {
			ret = append(ret, target)
		}
	}
	return ret, nil
}

func readXlsxSharedStrings(file *zip.File) ([]string, error) {
	if file == nil {
		return nil, nil
	}
	rc, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file.Name, err)
	}
	defer rc.Close()

	dec := xml.NewDecoder(io.LimitReader(rc, officeXMLPartMaxBytes))
	var ret []string
	var current strings.Builder
	inString := false
	inText := false
	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("parse %s: %w", file.Name, err)
		}

		switch value := token.(type) {
		case xml.StartElement:
			switch value.Name.Local {
			case "si":
				inString = true
				current.Reset()
			case "t":
				if inString {
					inText = true
				}
			}
		case xml.CharData:
			if inText {
				current.Write(value)
			}
		case xml.EndElement:
			switch value.Name.Local {
			case "t":
				inText = false
			case "si":
				ret = append(ret, current.String())
				inString = false
			}
		}
	}
	return ret, nil
}

func readXlsxSheetText(file *zip.File, sharedStrings []string) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open %s: %w", file.Name, err)
	}
	defer rc.Close()

	dec := xml.NewDecoder(io.LimitReader(rc, officeXMLPartMaxBytes))
	var ret strings.Builder
	cellType := ""
	inCell := false
	inValue := false
	inInlineText := false
	var value strings.Builder
	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("parse %s: %w", file.Name, err)
		}

		switch xmlToken := token.(type) {
		case xml.StartElement:
			switch xmlToken.Name.Local {
			case "c":
				inCell = true
				cellType = attrValue(xmlToken, "t")
				value.Reset()
			case "v":
				if inCell {
					inValue = true
				}
			case "t":
				if inCell && cellType == "inlineStr" {
					inInlineText = true
				}
			}
		case xml.CharData:
			if inValue || inInlineText {
				value.Write(xmlToken)
			}
		case xml.EndElement:
			switch xmlToken.Name.Local {
			case "v":
				inValue = false
			case "t":
				inInlineText = false
			case "c":
				appendXlsxCellText(&ret, cellType, strings.TrimSpace(value.String()), sharedStrings)
				inCell = false
				cellType = ""
			}
		}
	}
	return ret.String(), nil
}

func appendXlsxCellText(out *strings.Builder, cellType, value string, sharedStrings []string) {
	if value == "" {
		return
	}

	text := value
	if cellType == "s" {
		index, err := strconv.Atoi(value)
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return
		}
		text = sharedStrings[index]
	}
	if text == "" {
		return
	}
	out.WriteString(text)
	out.WriteByte(' ')
}

func decodeZipXML(file *zip.File, out any) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("open %s: %w", file.Name, err)
	}
	defer rc.Close()

	if err := xml.NewDecoder(io.LimitReader(rc, officeXMLPartMaxBytes)).Decode(out); err != nil {
		return fmt.Errorf("parse %s: %w", file.Name, err)
	}
	return nil
}

func attrValue(element xml.StartElement, name string) string {
	for _, attr := range element.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
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
