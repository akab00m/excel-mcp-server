package excel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Excel interface {
	// GetBackendName returns the backend used to manipulate the Excel file.
	GetBackendName() string
	// GetSheets returns a list of all worksheets in the Excel file.
	GetSheets() ([]Worksheet, error)
	// FindSheet finds a sheet by its name and returns a Worksheet.
	FindSheet(sheetName string) (Worksheet, error)
	// CreateNewSheet creates a new sheet with the specified name.
	CreateNewSheet(sheetName string) error
	// CopySheet copies a sheet from one to another.
	CopySheet(srcSheetName, destSheetName string) error
	// Save saves the Excel file.
	Save() error
}

type Worksheet interface {
	// Release releases the worksheet resources.
	Release()
	// Name returns the name of the worksheet.
	Name() (string, error)
	// GetTable returns a tables in this worksheet.
	GetTables() ([]Table, error)
	// GetPivotTable returns a pivot tables in this worksheet.
	GetPivotTables() ([]PivotTable, error)
	// SetValue sets a value in the specified cell.
	SetValue(cell string, value any) error
	// SetFormula sets a formula in the specified cell.
	SetFormula(cell string, formula string) error
	// GetValue gets the value from the specified cell.
	GetValue(cell string) (string, error)
	// GetFormula gets the formula from the specified cell.
	GetFormula(cell string) (string, error)
	// GetDimention gets the dimension of the worksheet.
	GetDimention() (string, error)
	// GetPagingStrategy returns the paging strategy for the worksheet.
	// The pageSize parameter is used to determine the max size of each page.
	GetPagingStrategy(pageSize int) (PagingStrategy, error)
	// CapturePicture returns base64 encoded image data of the specified range.
	CapturePicture(captureRange string) (string, error)
	// AddTable adds a table to this worksheet.
	AddTable(tableRange, tableName string) error
	// GetCellStyle gets style information for the specified cell.
	GetCellStyle(cell string) (*CellStyle, error)
	// SetCellStyle sets style for the specified cell.
	SetCellStyle(cell string, style *CellStyle) error
}

type Table struct {
	Name  string
	Range string
}

type PivotTable struct {
	Name  string
	Range string
}

type CellStyle struct {
	Border        []Border   `yaml:"border,omitempty"`
	Font          *FontStyle `yaml:"font,omitempty"`
	Fill          *FillStyle `yaml:"fill,omitempty"`
	NumFmt        *string    `yaml:"numFmt,omitempty"`
	DecimalPlaces *int       `yaml:"decimalPlaces,omitempty"`
}

type Border struct {
	Type  BorderType  `yaml:"type"`
	Style BorderStyle `yaml:"style,omitempty"`
	Color string      `yaml:"color,omitempty"`
}

type FontStyle struct {
	Bold      *bool          `yaml:"bold,omitempty"`
	Italic    *bool          `yaml:"italic,omitempty"`
	Underline *FontUnderline `yaml:"underline,omitempty"`
	Size      *int           `yaml:"size,omitempty"`
	Strike    *bool          `yaml:"strike,omitempty"`
	Color     *string        `yaml:"color,omitempty"`
	VertAlign *FontVertAlign `yaml:"vertAlign,omitempty"`
}

type FillStyle struct {
	Type    FillType     `yaml:"type,omitempty"`
	Pattern FillPattern  `yaml:"pattern,omitempty"`
	Color   []string     `yaml:"color,omitempty"`
	Shading *FillShading `yaml:"shading,omitempty"`
}

// OpenFile opens an Excel file and returns an Excel interface.
// It first tries to open the file using OLE automation, and if that fails,
// it tries to using the excelize library.
func OpenFile(absoluteFilePath string) (Excel, func(), error) {
	ole, releaseFn, err := NewExcelOle(absoluteFilePath)
	if err == nil {
		return ole, releaseFn, nil
	}
	// If OLE fails, try Excelize
	workbook, err := excelize.OpenFile(absoluteFilePath)
	if err != nil {
		return nil, func() {}, wrapOpenError(absoluteFilePath, err)
	}
	backend := NewExcelizeExcel(workbook)
	return backend, func() {
		workbook.Close()
	}, nil
}

// CreateFile creates a new empty workbook at absoluteFilePath.
// Parent directories are created as needed. Fails if the path already exists.
// If sheetName is non-empty, the first sheet is renamed to sheetName.
func CreateFile(absoluteFilePath string, sheetName string) error {
	if !filepath.IsAbs(absoluteFilePath) {
		return fmt.Errorf("path %q is not absolute", absoluteFilePath)
	}
	if _, err := os.Stat(absoluteFilePath); err == nil {
		return fmt.Errorf("file already exists: %s", absoluteFilePath)
	} else if !os.IsNotExist(err) {
		return err
	}
	dir := filepath.Dir(absoluteFilePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent directory %q: %w", dir, err)
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if sheetName != "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return fmt.Errorf("new workbook has no sheets")
		}
		if err := f.SetSheetName(sheets[0], sheetName); err != nil {
			return fmt.Errorf("set sheet name %q: %w", sheetName, err)
		}
	}
	if err := f.SaveAs(absoluteFilePath); err != nil {
		return fmt.Errorf("create workbook %q: %w", absoluteFilePath, err)
	}
	return nil
}

// OpenFileOrCreate opens absoluteFilePath, creating an empty workbook first if missing.
func OpenFileOrCreate(absoluteFilePath string) (Excel, func(), error) {
	_, err := os.Stat(absoluteFilePath)
	if os.IsNotExist(err) {
		if createErr := CreateFile(absoluteFilePath, ""); createErr != nil {
			return nil, func() {}, createErr
		}
	} else if err != nil {
		return nil, func() {}, err
	}
	return OpenFile(absoluteFilePath)
}

func wrapOpenError(path string, err error) error {
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: path %q not found; mount the same absolute path into this process as the agent uses (shared volume), e.g. /data/book.xlsx", err, path)
	}
	// excelize may return a plain string error; still hint on common missing-file text.
	msg := err.Error()
	if strings.Contains(msg, "no such file") || strings.Contains(msg, "cannot find the file") {
		return fmt.Errorf("%w: path %q not found; mount the same absolute path into this process as the agent uses (shared volume), e.g. /data/book.xlsx", err, path)
	}
	return err
}

// BorderType represents border direction
type BorderType string

const (
	BorderTypeLeft         BorderType = "left"
	BorderTypeRight        BorderType = "right"
	BorderTypeTop          BorderType = "top"
	BorderTypeBottom       BorderType = "bottom"
	BorderTypeDiagonalDown BorderType = "diagonalDown"
	BorderTypeDiagonalUp   BorderType = "diagonalUp"
)

func (b BorderType) String() string {
	return string(b)
}

func (b BorderType) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func BorderTypeValues() []BorderType {
	return []BorderType{
		BorderTypeLeft,
		BorderTypeRight,
		BorderTypeTop,
		BorderTypeBottom,
		BorderTypeDiagonalDown,
		BorderTypeDiagonalUp,
	}
}

// BorderStyle represents border style constants
type BorderStyle string

const (
	BorderStyleNone             BorderStyle = "none"
	BorderStyleContinuous       BorderStyle = "continuous"
	BorderStyleDash             BorderStyle = "dash"
	BorderStyleDot              BorderStyle = "dot"
	BorderStyleDouble           BorderStyle = "double"
	BorderStyleDashDot          BorderStyle = "dashDot"
	BorderStyleDashDotDot       BorderStyle = "dashDotDot"
	BorderStyleSlantDashDot     BorderStyle = "slantDashDot"
	BorderStyleMediumDashDot    BorderStyle = "mediumDashDot"
	BorderStyleMediumDashDotDot BorderStyle = "mediumDashDotDot"
)

func (b BorderStyle) String() string {
	return string(b)
}

func (b BorderStyle) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func BorderStyleValues() []BorderStyle {
	return []BorderStyle{
		BorderStyleNone,
		BorderStyleContinuous,
		BorderStyleDash,
		BorderStyleDot,
		BorderStyleDouble,
		BorderStyleDashDot,
		BorderStyleDashDotDot,
		BorderStyleSlantDashDot,
		BorderStyleMediumDashDot,
		BorderStyleMediumDashDotDot,
	}
}

// FontUnderline represents underline styles for font
type FontUnderline string

const (
	FontUnderlineNone             FontUnderline = "none"
	FontUnderlineSingle           FontUnderline = "single"
	FontUnderlineDouble           FontUnderline = "double"
	FontUnderlineSingleAccounting FontUnderline = "singleAccounting"
	FontUnderlineDoubleAccounting FontUnderline = "doubleAccounting"
)

func (f FontUnderline) String() string {
	return string(f)
}
func (f FontUnderline) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func FontUnderlineValues() []FontUnderline {
	return []FontUnderline{
		FontUnderlineNone,
		FontUnderlineSingle,
		FontUnderlineDouble,
		FontUnderlineSingleAccounting,
		FontUnderlineDoubleAccounting,
	}
}

// FontVertAlign represents vertical alignment options for font styles
type FontVertAlign string

const (
	FontVertAlignBaseline    FontVertAlign = "baseline"
	FontVertAlignSuperscript FontVertAlign = "superscript"
	FontVertAlignSubscript   FontVertAlign = "subscript"
)

func (v FontVertAlign) String() string {
	return string(v)
}

func (v FontVertAlign) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func FontVertAlignValues() []FontVertAlign {
	return []FontVertAlign{
		FontVertAlignBaseline,
		FontVertAlignSuperscript,
		FontVertAlignSubscript,
	}
}

// FillType represents fill types for cell styles
type FillType string

const (
	FillTypeGradient FillType = "gradient"
	FillTypePattern  FillType = "pattern"
)

func (f FillType) String() string {
	return string(f)
}

func (f FillType) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func FillTypeValues() []FillType {
	return []FillType{
		FillTypeGradient,
		FillTypePattern,
	}
}

// FillPattern represents fill pattern constants
type FillPattern string

const (
	FillPatternNone            FillPattern = "none"
	FillPatternSolid           FillPattern = "solid"
	FillPatternMediumGray      FillPattern = "mediumGray"
	FillPatternDarkGray        FillPattern = "darkGray"
	FillPatternLightGray       FillPattern = "lightGray"
	FillPatternDarkHorizontal  FillPattern = "darkHorizontal"
	FillPatternDarkVertical    FillPattern = "darkVertical"
	FillPatternDarkDown        FillPattern = "darkDown"
	FillPatternDarkUp          FillPattern = "darkUp"
	FillPatternDarkGrid        FillPattern = "darkGrid"
	FillPatternDarkTrellis     FillPattern = "darkTrellis"
	FillPatternLightHorizontal FillPattern = "lightHorizontal"
	FillPatternLightVertical   FillPattern = "lightVertical"
	FillPatternLightDown       FillPattern = "lightDown"
	FillPatternLightUp         FillPattern = "lightUp"
	FillPatternLightGrid       FillPattern = "lightGrid"
	FillPatternLightTrellis    FillPattern = "lightTrellis"
	FillPatternGray125         FillPattern = "gray125"
	FillPatternGray0625        FillPattern = "gray0625"
)

func (f FillPattern) String() string {
	return string(f)
}

func (f FillPattern) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func FillPatternValues() []FillPattern {
	return []FillPattern{
		FillPatternNone,
		FillPatternSolid,
		FillPatternMediumGray,
		FillPatternDarkGray,
		FillPatternLightGray,
		FillPatternDarkHorizontal,
		FillPatternDarkVertical,
		FillPatternDarkDown,
		FillPatternDarkUp,
		FillPatternDarkGrid,
		FillPatternDarkTrellis,
		FillPatternLightHorizontal,
		FillPatternLightVertical,
		FillPatternLightDown,
		FillPatternLightUp,
		FillPatternLightGrid,
		FillPatternLightTrellis,
		FillPatternGray125,
		FillPatternGray0625,
	}
}

// FillShading represents fill shading constants
type FillShading string

const (
	FillShadingHorizontal   FillShading = "horizontal"
	FillShadingVertical     FillShading = "vertical"
	FillShadingDiagonalDown FillShading = "diagonalDown"
	FillShadingDiagonalUp   FillShading = "diagonalUp"
	FillShadingFromCenter   FillShading = "fromCenter"
	FillShadingFromCorner   FillShading = "fromCorner"
)

func (f FillShading) String() string {
	return string(f)
}

func (f FillShading) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func FillShadingValues() []FillShading {
	return []FillShading{
		FillShadingHorizontal,
		FillShadingVertical,
		FillShadingDiagonalDown,
		FillShadingDiagonalUp,
		FillShadingFromCenter,
		FillShadingFromCorner,
	}
}
