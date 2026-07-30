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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDocxText(t *testing.T) {
	path := writeOfficeTestArchive(t, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
	<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
	<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
	<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>
</Types>`,
		"word/header1.xml":  `<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:p><w:r><w:t>Header Text</w:t></w:r></w:p></w:hdr>`,
		"word/document.xml": `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Hello</w:t><w:tab/><w:t>DOCX</w:t></w:r></w:p><w:p><w:instrText>skip me</w:instrText><w:r><w:t>Body</w:t></w:r></w:p></w:body></w:document>`,
		"word/footer1.xml":  `<w:ftr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:p><w:r><w:t>Footer Text</w:t></w:r></w:p></w:ftr>`,
	})

	text, err := extractDocxText(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Header Text") || !strings.Contains(text, "Hello") || !strings.Contains(text, "DOCX") || !strings.Contains(text, "Body") || !strings.Contains(text, "Footer Text") {
		t.Fatalf("missing extracted DOCX content: %q", text)
	}
	if strings.Contains(text, "skip me") {
		t.Fatalf("field instruction text should be skipped: %q", text)
	}
}

func TestExtractPptxText(t *testing.T) {
	path := writeOfficeTestArchive(t, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
	<Override PartName="/ppt/slides/slide1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
	<Override PartName="/ppt/diagrams/data1.xml" ContentType="application/vnd.openxmlformats-officedocument.drawingml.diagramData+xml"/>
</Types>`,
		"ppt/slides/slide1.xml":  `<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:t>Slide One</a:t></p:sld>`,
		"ppt/diagrams/data1.xml": `<dgm:dataModel xmlns:dgm="http://schemas.openxmlformats.org/drawingml/2006/diagram"><dgm:t>Diagram Text</dgm:t></dgm:dataModel>`,
	})

	text, err := extractPptxText(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Slide One") || !strings.Contains(text, "Diagram Text") {
		t.Fatalf("missing extracted PPTX content: %q", text)
	}
}

func writeOfficeTestArchive(t *testing.T, files map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "office.zip")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	zw := zip.NewWriter(out)
	for name, content := range files {
		w, createErr := zw.Create(name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := w.Write([]byte(content)); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
