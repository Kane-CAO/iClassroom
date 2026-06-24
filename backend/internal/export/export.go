package exportutil

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

// SubmissionRow is one Excel row in the exported workbook.
type SubmissionRow struct {
	RoomCode        string
	RoomTitle       string
	GroupName       string
	StudentNickname string
	TaskTitle       string
	ContentText     string
	ImageFileNames  string
	FileNames       string
	Score           string
	Comment         string
	SubmittedAt     string
	GradedAt        string
}

// ImageFile is one image entry to embed into the zip archive.
type ImageFile struct {
	ArchivePath string
	FilePath    string
}

type AttachmentFile struct {
	ArchivePath string
	FilePath    string
}

// BuildSubmissionsArchive returns a zip archive containing submissions.xlsx
// plus all supplied images.
func BuildSubmissionsArchive(rows []SubmissionRow, images []ImageFile) ([]byte, error) {
	return BuildSubmissionsArchiveWithFiles(rows, images, nil)
}

func BuildSubmissionsArchiveWithFiles(rows []SubmissionRow, images []ImageFile, files []AttachmentFile) ([]byte, error) {
	xlsx, err := buildWorkbook(rows)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := writeZipEntry(zw, "submissions.xlsx", xlsx); err != nil {
		_ = zw.Close()
		return nil, err
	}

	for _, image := range images {
		data, err := os.ReadFile(image.FilePath)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("exportutil: read image file %q: %w", image.FilePath, err)
		}
		archivePath := strings.TrimPrefix(path.Clean("/"+image.ArchivePath), "/")
		if archivePath == "." || archivePath == "" {
			_ = zw.Close()
			return nil, fmt.Errorf("exportutil: invalid archive path %q", image.ArchivePath)
		}
		if err := writeZipEntry(zw, path.Join("images", archivePath), data); err != nil {
			_ = zw.Close()
			return nil, err
		}
	}

	for _, file := range files {
		data, err := os.ReadFile(file.FilePath)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("exportutil: read attachment file %q: %w", file.FilePath, err)
		}
		archivePath := strings.TrimPrefix(path.Clean("/"+file.ArchivePath), "/")
		if archivePath == "." || archivePath == "" {
			_ = zw.Close()
			return nil, fmt.Errorf("exportutil: invalid archive path %q", file.ArchivePath)
		}
		if err := writeZipEntry(zw, path.Join("files", archivePath), data); err != nil {
			_ = zw.Close()
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("exportutil: close zip: %w", err)
	}

	return buf.Bytes(), nil
}

func writeZipEntry(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("exportutil: create zip entry %q: %w", name, err)
	}
	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("exportutil: write zip entry %q: %w", name, err)
	}
	return nil
}

func buildWorkbook(rows []SubmissionRow) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := writeZipEntry(zw, "[Content_Types].xml", []byte(contentTypesXML)); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeZipEntry(zw, "_rels/.rels", []byte(rootRelsXML)); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeZipEntry(zw, "xl/workbook.xml", []byte(workbookXML)); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeZipEntry(zw, "xl/_rels/workbook.xml.rels", []byte(workbookRelsXML)); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeZipEntry(zw, "xl/styles.xml", []byte(stylesXML)); err != nil {
		_ = zw.Close()
		return nil, err
	}
	sheetXML, err := buildSheetXML(rows)
	if err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeZipEntry(zw, "xl/worksheets/sheet1.xml", sheetXML); err != nil {
		_ = zw.Close()
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("exportutil: close xlsx zip: %w", err)
	}

	return buf.Bytes(), nil
}

func buildSheetXML(rows []SubmissionRow) ([]byte, error) {
	var b strings.Builder
	lastRow := len(rows) + 1
	if lastRow < 1 {
		lastRow = 1
	}

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	b.WriteString(`<dimension ref="A1:L`)
	b.WriteString(strconv.Itoa(lastRow))
	b.WriteString(`"/>`)
	b.WriteString(`<sheetData>`)

	writeExcelRow(&b, 1, []string{
		"roomCode",
		"roomTitle",
		"groupName",
		"studentNickname",
		"taskTitle",
		"contentText",
		"imageFileNames",
		"fileNames",
		"score",
		"comment",
		"submittedAt",
		"gradedAt",
	})

	for i, row := range rows {
		writeExcelRow(&b, i+2, []string{
			row.RoomCode,
			row.RoomTitle,
			row.GroupName,
			row.StudentNickname,
			row.TaskTitle,
			row.ContentText,
			row.ImageFileNames,
			row.FileNames,
			row.Score,
			row.Comment,
			row.SubmittedAt,
			row.GradedAt,
		})
	}

	b.WriteString(`</sheetData>`)
	b.WriteString(`<pageMargins left="0.7" right="0.7" top="0.75" bottom="0.75" header="0.3" footer="0.3"/>`)
	b.WriteString(`</worksheet>`)

	return []byte(b.String()), nil
}

func writeExcelRow(b *strings.Builder, rowIndex int, values []string) {
	b.WriteString(`<row r="`)
	b.WriteString(strconv.Itoa(rowIndex))
	b.WriteString(`">`)
	for i, value := range values {
		b.WriteString(`<c r="`)
		b.WriteString(excelColumnName(i + 1))
		b.WriteString(strconv.Itoa(rowIndex))
		b.WriteString(`" t="inlineStr"><is><t xml:space="preserve">`)
		b.WriteString(escapeXML(value))
		b.WriteString(`</t></is></c>`)
	}
	b.WriteString(`</row>`)
}

func excelColumnName(index int) string {
	if index <= 0 {
		return ""
	}
	var chars []byte
	for index > 0 {
		index--
		chars = append([]byte{byte('A' + index%26)}, chars...)
		index /= 26
	}
	return string(chars)
}

func escapeXML(value string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const workbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Submissions" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`

const workbookRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="1">
    <font>
      <sz val="11"/>
      <color theme="1"/>
      <name val="Calibri"/>
      <family val="2"/>
    </font>
  </fonts>
  <fills count="1">
    <fill><patternFill patternType="none"/></fill>
  </fills>
  <borders count="1">
    <border>
      <left/><right/><top/><bottom/><diagonal/>
    </border>
  </borders>
  <cellStyleXfs count="1">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0"/>
  </cellStyleXfs>
  <cellXfs count="1">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
  </cellXfs>
</styleSheet>`
