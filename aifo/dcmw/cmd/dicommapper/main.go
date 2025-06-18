// cmd/dicommapper/main.go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"text/template"
)

type DateTimeFieldInfo struct {
	Name     string
	Group    string
	Element  string
	DateName string // Name of the date field
	TimeName string // Name of the time field
}

type FieldInfo struct {
	Name    string
	Type    string
	Group   string
	Element string
}

type StructInfo struct {
	Name           string
	RegularFields  []FieldInfo
	DateTimeFields []DateTimeFieldInfo
	NeedsElem      bool
	NeedsErr       bool
}

func main() {
	// Parse the models package
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "../models", nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing package:", err)
		os.Exit(1)
	}

	var structs []StructInfo

	// Iterate over the packages and files
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			// Iterate over the declarations
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

					structInfo := StructInfo{
						Name: typeSpec.Name.Name,
					}

					// Iterate over the fields
					for _, field := range structType.Fields.List {
						if field.Tag == nil {
							continue
						}
						tagValue := field.Tag.Value // This includes the backticks
						tagValue = strings.Trim(tagValue, "`")
						structTag := reflect.StructTag(tagValue)

						// Parse the DICOM tag
						dicomTagStr := structTag.Get("dicom")
						// Parse the datetime tag
						datetimeTagStr := structTag.Get("datetime")

						// If both dicom and datetime are missing, skip this field
						if dicomTagStr == "" && datetimeTagStr == "" {
							continue
						}

						fieldType := getFieldType(field.Type)

						// Handle the datetime tag if present
						if datetimeTagStr != "" {
							// Split the datetime tag into two parts: the date and time fields
							datetimeParts := strings.Split(datetimeTagStr, "-")
							if len(datetimeParts) != 2 {
								fmt.Printf("Invalid datetime format for field %s: %s\n", field.Names[0].Name, datetimeTagStr)
								continue
							}

							var datetimeInfo DateTimeFieldInfo
							var valid = true

							// Find and validate date field
							dateFieldName := datetimeParts[0]
							dateField, found := findFieldByName(structType, dateFieldName)
							if !found || dateField.Tag == nil {
								fmt.Printf("Missing date field %s for datetime field %s\n", dateFieldName, field.Names[0].Name)
								valid = false
							} else {
								dateDicomTag := extractDicomTag(dateField.Tag.Value)
								if dateDicomTag == nil {
									fmt.Printf("Missing dicom tag for date field %s required by datetime field %s\n", dateFieldName, field.Names[0].Name)
									valid = false
								} else {
									datetimeInfo.DateName = dateFieldName
									datetimeInfo.Group = dateDicomTag[0]
									datetimeInfo.Element = dateDicomTag[1]
								}
							}

							// Find and validate time field
							timeFieldName := datetimeParts[1]
							timeField, found := findFieldByName(structType, timeFieldName)
							if !found || timeField.Tag == nil {
								fmt.Printf("Missing time field %s for datetime field %s\n", timeFieldName, field.Names[0].Name)
								valid = false
							} else {
								timeDicomTag := extractDicomTag(timeField.Tag.Value)
								if timeDicomTag == nil {
									fmt.Printf("Missing dicom tag for time field %s required by datetime field %s\n", timeFieldName, field.Names[0].Name)
									valid = false
								} else {
									datetimeInfo.TimeName = timeFieldName
									// Assuming date and time fields have the same group
									if datetimeInfo.Group == "" {
										datetimeInfo.Group = timeDicomTag[0]
									}
									datetimeInfo.Element = timeDicomTag[1]
								}
							}

							// Only add if both date and time fields are valid
							if valid {
								datetimeInfo.Name = field.Names[0].Name
								structInfo.DateTimeFields = append(structInfo.DateTimeFields, datetimeInfo)
								structInfo.NeedsElem = true
								structInfo.NeedsErr = true
							}
							continue
						}

						// Handle the regular dicom tag
						if dicomTagStr != "" {
							// Split the DICOM tag and ensure we have exactly two parts
							dicomTagParts := strings.Split(dicomTagStr, ",")
							if len(dicomTagParts) != 2 {
								fmt.Printf("Invalid DICOM tag format for field %s: %s\n", field.Names[0].Name, dicomTagStr)
								continue
							}

							fieldInfo := FieldInfo{
								Name:    field.Names[0].Name,
								Type:    fieldType,
								Group:   dicomTagParts[0],
								Element: dicomTagParts[1],
							}

							// Add to RegularFields
							structInfo.RegularFields = append(structInfo.RegularFields, fieldInfo)
							structInfo.NeedsElem = true
							structInfo.NeedsErr = true
						}
					}

					// Only add structs that have fields with DICOM tags or datetime fields
					if len(structInfo.RegularFields) > 0 || len(structInfo.DateTimeFields) > 0 {
						structs = append(structs, structInfo)
					}
				}
			}
		}
	}

	// Parse the template from file
	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
	}

	tmpl, err := template.New("dicom_mapping.tmpl").Funcs(funcMap).ParseFiles("../processor/dicom_mapping.tmpl")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		os.Exit(1)
	}

	// Create the output file
	outputFile, err := os.Create("../processor/dicom_mapping_generated.go")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	// Execute the template
	err = tmpl.Execute(outputFile, structs)
	if err != nil {
		fmt.Println("Error executing template:", err)
		os.Exit(1)
	}

	fmt.Println("Code generation for dicom_mapping_generated.go complete.")
}

// Helper function to find a field by name in a struct
func findFieldByName(structType *ast.StructType, name string) (*ast.Field, bool) {
	for _, field := range structType.Fields.List {
		if field.Names != nil && len(field.Names) > 0 && field.Names[0].Name == name {
			return field, true
		}
	}
	return nil, false
}

// Helper function to extract DICOM tag from a field tag
func extractDicomTag(tagValue string) []string {
	tagValue = strings.Trim(tagValue, "`")
	structTag := reflect.StructTag(tagValue)
	dicomTagStr := structTag.Get("dicom")
	if dicomTagStr == "" {
		return nil
	}
	dicomTagParts := strings.Split(dicomTagStr, ",")
	if len(dicomTagParts) != 2 {
		return nil
	}
	return dicomTagParts
}

// getFieldType returns the string representation of the field type
func getFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getFieldType(t.X)
	case *ast.SelectorExpr:
		pkgIdent, ok := t.X.(*ast.Ident)
		if ok {
			return pkgIdent.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	default:
		return ""
	}
}
