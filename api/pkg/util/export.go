package util

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unicode"

	"github.com/xuri/excelize/v2"
)

type ExportField struct {
	FieldName string
	CnName    string
	FieldType string
	Selects   map[string]string
}
type ExportStruct struct {
}

var globalMu sync.Mutex
var exportStructInstance *ExportStruct

func ExportXlsx[T any](fields []ExportField, w io.Writer, dataFn func() ([]T, bool, error)) error {
	if exportStructInstance == nil {
		globalMu.Lock()
		defer globalMu.Unlock()
		// 获取锁后再次检查，防止多个goroutine同时通过第一次检查后重复创建
		if exportStructInstance == nil {
			exportStructInstance = &ExportStruct{}
		}
	}
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// 1. 初始化一个切片用于记录每列的最大宽度（单位：字符宽度近似值）
	// 初始值可以设为0，也可以设为表头宽度，这里我们初始化为0
	maxColWidths := make([]float64, len(fields))
	getTitle := func(i int) string {
		var result string
		for i >= 0 {
			remainder := i % 26
			result = string(rune('A'+remainder)) + result
			i = i/26 - 1
		}
		return result
	}
	// 2. 动态计算表头宽度并更新maxColWidths
	for i, field := range fields {
		cell := getTitle(i) + "1"
		f.SetCellValue(sheetName, cell, field.CnName)

		// 计算表头字符串的近似宽度，中文字符通常占2个宽度，英文字符占1个
		headerWidth := exportStructInstance.calculateStringWidth(field.CnName)
		if headerWidth > maxColWidths[i] {
			maxColWidths[i] = headerWidth
		}
	}

	// 3. 定义样式（边框和居中）
	styleID, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	rowIndex := 2
	for {
		rows, ok, err := dataFn()
		if err != nil {
			return err
		}

		for _, rowData := range rows {
			// 4. 遍历每一行数据，计算每个单元格的宽度并更新maxColWidths
			for i, field := range fields {
				v, err := exportStructInstance.getFieldValueString(rowData, field.FieldName)
				if err != nil {
					return err
				}
				cell := getTitle(i) + strconv.Itoa(rowIndex)
				f.SetCellValue(sheetName, cell, v)

				// 计算当前单元格内容的近似宽度
				contentWidth := exportStructInstance.calculateStringWidth(v)
				// 留大者：比较并更新最大宽度
				if contentWidth > maxColWidths[i] {
					maxColWidths[i] = contentWidth
				}
			}
			rowIndex++
		}

		if !ok {
			break
		}
	}

	// 5. 所有数据写入完成后，根据maxColWidths设置列宽
	for i, width := range maxColWidths {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		// 可以添加一个额外的缓冲值（例如2）让表格看起来更宽松
		finalWidth := width + 2
		f.SetColWidth(sheetName, colName, colName, finalWidth)
	}

	// 6. 设置行高20磅
	targetRowHeight := 20.0
	for r := 1; r < rowIndex; r++ { // rowIndex 现在是最后一行的行号+1
		f.SetRowHeight(sheetName, r, targetRowHeight)
	}

	// 7. 应用样式到整个数据区域
	startCell, _ := excelize.CoordinatesToCellName(1, 1)
	endCell, _ := excelize.CoordinatesToCellName(len(fields), rowIndex-1)
	f.SetCellStyle(sheetName, startCell, endCell, styleID)

	f.SetActiveSheet(index)
	return f.Write(w)
}

func (e *ExportStruct) calculateStringWidth(s string) float64 {
	width := 0.0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) { // 中文字符
			width += 2
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) { // 英文字母或数字
			width += 1
		} else if unicode.IsPunct(r) || unicode.IsSymbol(r) { // 标点或符号
			width += 1
		} else { // 其他字符（如空格、制表符等），可以赋予一个默认宽度
			width += 1
		}
	}
	return width
}

func (e *ExportStruct) getFieldValueString(u interface{}, fieldName string) (string, error) {
	v := reflect.ValueOf(u)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected a struct, got %s", v.Kind())
	}
	fieldVal := v.FieldByName(fieldName)
	if !fieldVal.IsValid() {
		return "", fmt.Errorf("field '%s' not found", fieldName)
	}
	if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
		timeVal := fieldVal.Interface().(time.Time)
		return timeVal.Format("2006-01-02 15:04:05"), nil
	}
	return fmt.Sprintf("%v", fieldVal.Interface()), nil
}
