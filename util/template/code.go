package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type ModelValue struct {
	Name          string       //模型名称
	LowerName     string       //模型名称小写
	Model         string       //模型中文名
	Fields        []FieldValue //字段
	FirstStrField *FieldValue  //第一个字符串字段
	PackageName   string
}
type FieldValue struct {
	FieldName      string //字段名称
	FieldLowerName string //字段名称小写
	CnName         string //字段中文名
	LangKey        string //字段语言key
	IsGorm         bool   //是否gorm字段
	IsNormal       bool   //是否普通字段
	ColumnName     string //数据库字段名
	FieldType      string //字段类型 input datetime select
	Select         map[string]string
}

//go:embed *.txt
var localeFS embed.FS

func main() {
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)
	mainPath := filepath.Join(dir, "..")
	list := []string{
		"data",
		"dead_msg",
		"log",
		"offline_chat_msg",
		"plugin",
		"plugin_column",
		"plugin_feature",
		"plugin_table",
		"plugin_web_data",
		"user",
		"user_settings",
	}
	menu := "user"
	list = []string{
		"desktop_app",
	}
	for _, model := range list {
		createCode(model, mainPath, menu)
	}
}

func camelToSnake(str string) string {
	// 匹配所有大写字母，并在其前添加下划线（首字符除外）
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
func parseModelFile(filePath string) (ModelValue, error) {
	result := ModelValue{}
	fset := token.NewFileSet()
	src, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("读取文件失败: %v", err)
	}
	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return result, fmt.Errorf("解析文件失败: %v", err)
	}
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			result.Name = typeSpec.Name.Name
			result.LowerName = camelToSnake(result.Name)
			result.Model = typeSpec.Name.Name
			if genDecl.Doc != nil {
				for _, comment := range genDecl.Doc.List {
					text := comment.Text
					if strings.Contains(text, "@model") {
						result.Model = extractValue(text, "name")
					}
				}
			}
			gormFields := []FieldValue{}
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					isGorm := false
					if se, ok := field.Type.(*ast.SelectorExpr); ok {
						if xIdent, ok := se.X.(*ast.Ident); ok {
							fmt.Println("f1", xIdent.Name, se.Sel.Name)
							if xIdent.Name == "base" && se.Sel.Name == "BaseModel" {
								isGorm = true
							}
						}
					} else if ident, ok := field.Type.(*ast.Ident); ok {
						if ident.Name == "BaseModel" {
							isGorm = true
						}
					}
					if isGorm {
						gormFields = []FieldValue{
							{FieldName: "ID", FieldLowerName: "id", CnName: "ID", LangKey: "model.common.id", IsGorm: true, IsNormal: false, ColumnName: "id", FieldType: "input"},
							{FieldName: "CreatedAt", FieldLowerName: "createdat", CnName: "创建时间", LangKey: "model.common.createdat", IsGorm: true, IsNormal: false, ColumnName: "created_at", FieldType: "datetime"},
							{FieldName: "UpdatedAt", FieldLowerName: "updatedat", CnName: "更新时间", LangKey: "model.common.updatedat", IsGorm: true, IsNormal: false, ColumnName: "updated_at", FieldType: "datetime"},
						}
					}
					continue
				}
				for _, fieldName := range field.Names {
					fv := FieldValue{
						FieldName: fieldName.Name,
						IsGorm:    false,
						IsNormal:  true,
						FieldType: "input",
					}
					if field.Doc != nil {
						for _, comment := range field.Doc.List {
							text := comment.Text
							if strings.Contains(text, "@field") {
								fv.CnName = extractValue(text, "name")
								fv.FieldLowerName = camelToSnake(fieldName.Name)
								fv.LangKey = "model." + result.LowerName + "." + fv.FieldLowerName
							}
							if strings.Contains(text, "@select") {
								fv.FieldType = "select"
								fv.Select = make(map[string]string)
								tmp := strings.Split(text, "@select")[1]
								for _, line := range strings.Split(strings.TrimSpace(tmp), " ") {
									kv := strings.SplitN(line, "=", 2)
									if len(kv) == 2 {
										fv.Select[kv[0]] = kv[1]
									}
								}
							}
							if strings.Contains(text, "@datetime") {
								fv.FieldType = "datetime"
							}
							if strings.Contains(text, "@textarea") {
								fv.FieldType = "textarea"
							}
						}
					}
					if fv.CnName == "" {
						fv.CnName = fieldName.Name
					}
					if field.Tag != nil {
						rawTag := field.Tag.Value
						trimmedTag := strings.Trim(rawTag, "`")
						allTags := strings.Fields(trimmedTag)
						for _, tagPair := range allTags {
							parts := strings.SplitN(tagPair, ":", 2)
							if len(parts) == 2 && parts[0] == "gorm" {
								gormTagValue := parts[1]
								gormTagValue = strings.Trim(gormTagValue, `"`)
								gormAttrs := strings.Split(gormTagValue, ";")
								for _, attr := range gormAttrs {
									attr = strings.TrimSpace(attr)
									if strings.HasPrefix(attr, "column:") {
										fv.ColumnName = strings.TrimPrefix(attr, "column:")
										break
									}
								}
								break
							}
						}
					}
					if fv.ColumnName == "" {
						continue
					}
					if result.FirstStrField == nil {
						fieldType := field.Type
						if ident, ok := fieldType.(*ast.Ident); ok && ident.Name == "string" {
							result.FirstStrField = &fv
						}
					}
					result.Fields = append(result.Fields, fv)
				}
			}
			if len(gormFields) > 0 {
				result.Fields = append(result.Fields, gormFields...)
			}
			return result, nil
		}
	}
	return result, nil
}
func extractValue(commentText, key string) string {
	text := strings.TrimSpace(commentText)
	if after, ok := strings.CutPrefix(text, "//"); ok {
		text = after
	}
	parts := strings.Fields(text)
	for _, part := range parts {
		if strings.Contains(part, key+"=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				return strings.Trim(kv[1], `"`)
			}
		}
	}
	return ""
}

func createCode(model string, mainPath string, menu string) {
	modelPath := filepath.Join(mainPath, "api", "internal", "model", menu, model+".go")
	zhPath := filepath.Join(mainPath, "api", "pkg", "locales", "model-zh.json")
	enPath := filepath.Join(mainPath, "api", "pkg", "locales", "model-en.json")
	// 解析模型文件
	modelValue, err := parseModelFile(modelPath)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}
	modelValue.PackageName = menu
	//1.生成翻译
	createFy(modelValue, zhPath, false)
	println("中文翻译生成完成")
	createFy(modelValue, enPath, true)
	println("英文翻译生成完成")
	//3.生成控制器
	createTemplate(modelValue, "controller.txt", filepath.Join(mainPath, "api", "internal", "controller", "api", menu, modelValue.LowerName+".go"))
	println("控制器生成完成")
	//4.生成index页面
	createTemplate(modelValue, "index.txt", filepath.Join(mainPath, "web", "src", "components", "common", "views", menu, modelValue.LowerName, "index.vue"))
	println("index页面生成完成")
	//5.生成edit页面
	createTemplate(modelValue, "edit.txt", filepath.Join(mainPath, "web", "src", "components", "common", "views", menu, modelValue.LowerName, "edit.vue"))
	println("edit页面生成完成")
	//6.生成detail页面
	createTemplate(modelValue, "detail.txt", filepath.Join(mainPath, "web", "src", "components", "common", "views", menu, modelValue.LowerName, "detail.vue"))
	println("detail页面生成完成")
	//7.生成import页面
	createTemplate(modelValue, "import.txt", filepath.Join(mainPath, "web", "src", "components", "common", "views", menu, modelValue.LowerName, "import.vue"))
	fmt.Println("import页面生成完成")
}

func createTemplate(modelValue ModelValue, tmplFile, path string) {
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := localeFS.ReadFile(tmplFile)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}
	tmpl, err := template.New("test").Delims("[[", "]]").Parse(string(data))
	if err != nil {
		fmt.Printf("解析模板失败: %v\n", err)
		return
	}
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer f.Close()
	err = tmpl.Execute(f, modelValue)
	if err != nil {
		fmt.Printf("执行模板失败: %v\n", err)
		return
	}
	fmt.Printf("生成文件成功: %s\n", path)
}
func createFy(modelValue ModelValue, zhPath string, eng bool) {
	zhContent, err := os.ReadFile(zhPath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}
	var zhMap map[string]interface{}
	err = json.Unmarshal(zhContent, &zhMap)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}
	if zhMap["model"].(map[string]interface{})[modelValue.LowerName] == nil {
		zhMap["model"].(map[string]interface{})[modelValue.LowerName] = make(map[string]interface{})
	}
	data := zhMap["model"].(map[string]interface{})[modelValue.LowerName].(map[string]interface{})
	for _, field := range modelValue.Fields {
		if field.IsGorm {
			continue
		}
		if !eng || data[field.FieldLowerName] == nil {
			data[field.FieldLowerName] = field.CnName
		}
		if field.FieldType == "select" {
			for k, v := range field.Select {
				if !eng || data[field.FieldLowerName+"_select_"+k] == nil {
					data[field.FieldLowerName+"_select_"+k] = v
				}
			}
		}
	}
	if !eng || data["model"] == nil {
		data["model"] = modelValue.Model
	}
	zhContent, err = json.MarshalIndent(zhMap, "", "  ")
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}
	err = os.WriteFile(zhPath, zhContent, 0644)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}
}
