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
	"bytes"
	"compress/zlib"
	"strings"
	"testing"
)

func TestExtractPDFTextFromFlateTextStream(t *testing.T) {
	content := []byte("BT /F1 12 Tf 72 720 Td (Hello Scribli) Tj T* [(PDF) 120 (search)] TJ ET")
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	pdf := append([]byte("%PDF-1.7\n1 0 obj << /Type /Page >> endobj\n2 0 obj << /Length 1 /Filter /FlateDecode >>\nstream\n"), compressed.Bytes()...)
	pdf = append(pdf, []byte("\nendstream\nendobj\n%%EOF")...)

	text, err := extractPDFText(pdf)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Hello Scribli", "PDF", "search"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in extracted text %q", want, text)
		}
	}
}

func TestCountPDFPagesIgnoresPageTree(t *testing.T) {
	pdf := []byte("<< /Type /Pages /Count 2 >> << /Type /Page >> << /Type /Page >>")
	if got := countPDFPages(pdf); got != 2 {
		t.Fatalf("expected 2 pages, got %d", got)
	}
}

func TestExtractPDFTextRejectsEncryptedPDF(t *testing.T) {
	if _, err := extractPDFText([]byte("%PDF-1.7\n<< /Encrypt 1 0 R >>")); err == nil {
		t.Fatal("expected encrypted PDF to be rejected")
	}
}
