package parser

import (
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"strings"
)

// extractMethods extracts methods from an interface declaration
func extractMethods(interfaceType *ast.InterfaceType) []MethodInfo {
	var methods []MethodInfo
	for _, field := range interfaceType.Methods.List {
		if funcType, ok := field.Type.(*ast.FuncType); ok {
			methodInfo := MethodInfo{
				Name:       field.Names[0].Name,
				Parameters: extractParameters(funcType.Params),
				Returns:    extractParameters(funcType.Results),
			}
			methods = append(methods, methodInfo)
		}
	}
	return methods
}

// extractFields extracts fields from a struct
func extractFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo
	for _, field := range structType.Fields.List {
		typeStr := formatExpr(field.Type)
		for _, name := range field.Names {
			fieldInfo := FieldInfo{
				Name: name.Name,
				Type: typeStr,
				Tag:  extractTag(field),
			}
			fields = append(fields, fieldInfo)
		}
	}
	return fields
}

// extractTag extracts struct tags
func extractTag(field *ast.Field) string {
	if field.Tag != nil {
		return strings.Trim(field.Tag.Value, "`")
	}
	return ""
}

// DescriptionData contains different parts of a function's documentation comment
type DescriptionData struct {
	Description     string
	Example         string
	Notes           string
	DeprecationNote string

	// Raw fields
	DescriptionRaw     string
	DeprecationNoteRaw string
}

// extractDescriptionData extracts the description and example code from a
// function's documentation comment
func extractDescriptionData(doc string) DescriptionData {
	lines := strings.Split(doc, "\n")

	var descLines []string
	var exampleLines []string
	var notesLines []string
	var deprecationNoteLines []string

	var description string
	var example string
	var notes string
	var deprecationNote string

	isExample := false
	isNotes := false
	isDeprecationNote := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Example:") {
			isExample = true
			isNotes = false
			isDeprecationNote = false
			continue
		}
		if strings.HasPrefix(line, "Notes:") {
			isNotes = true
			isExample = false
			isDeprecationNote = false
			continue
		}
		if strings.HasPrefix(line, "Deprecated:") {
			isDeprecationNote = true
			isExample = false
			isNotes = false
			continue
		}

		if isExample {
			exampleLines = append(exampleLines, line)
		} else if isNotes {
			notesLines = append(notesLines, line)
		} else if isDeprecationNote {
			deprecationNoteLines = append(deprecationNoteLines, line)
		} else {
			descLines = append(descLines, line)
		}
	}

	// Description
	descriptionRaw := strings.Join(descLines, "\n")
	description = strings.Join(descLines, "</p>\n<p>")
	description = "<p>" + description + "</p>"
	description = strings.ReplaceAll(description, "\t", " ")
	if description == "<p></p>" {
		description = ""
	}

	// Example
	example = strings.Join(exampleLines, "\n")
	example = strings.TrimLeft(example, " \t")
	example = strings.TrimLeft(example, "\n")
	example = formatExample(example)

	// Notes
	notes = strings.Join(notesLines, "</p>\n<p>")
	notes = "<p>" + notes + "</p>"
	notes = strings.ReplaceAll(notes, "\t", " ")
	if notes == "<p></p>" {
		notes = ""
	}

	// Deprecation Note
	deprecationNoteRaw := strings.Join(deprecationNoteLines, "\n")
	deprecationNote = strings.Join(deprecationNoteLines, "</p>\n<p>")
	deprecationNote = "<p>" + deprecationNote + "</p>"
	deprecationNote = strings.ReplaceAll(deprecationNote, "\t", " ")
	if deprecationNote == "<p></p>" {
		deprecationNote = ""
	}

	return DescriptionData{
		Description:     description,
		Example:         example,
		Notes:           notes,
		DeprecationNote: deprecationNote,

		// Raw fields
		DescriptionRaw:     descriptionRaw,
		DeprecationNoteRaw: deprecationNoteRaw,
	}
}

// formatExample formats the example code using the go/format package
func formatExample(example string) string {
	src := []byte(example)
	formattedSrc, err := format.Source(src)
	if err != nil {
		return example
	}

	return string(formattedSrc)
}

// extractParameters extracts the parameters from a function or method declaration
func extractParameters(fieldList *ast.FieldList) []string {
	var params []string
	if fieldList != nil {
		for _, param := range fieldList.List {
			typeStr := formatExpr(param.Type)
			for _, name := range param.Names {
				params = append(params, name.Name+" "+typeStr)
			}
			if len(param.Names) == 0 {
				params = append(params, typeStr)
			}
		}
	}
	return params
}

// extractBody extracts the body of a function declaration
func extractBody(fs *token.FileSet, fn *ast.FuncDecl) string {
	if fn.Body == nil {
		return ""
	}
	start := fs.Position(fn.Body.Pos()).Offset
	end := fs.Position(fn.Body.End()).Offset
	fileContent, _ := os.ReadFile(fs.File(fn.Body.Pos()).Name())
	return string(fileContent[start:end])
}

// formatExpr formats an expression using the go/format package
func formatExpr(expr ast.Expr) string {
	var out strings.Builder
	if err := format.Node(&out, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return out.String()
}

// extractComment extracts the comment text from a comment group
func extractComment(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return strings.TrimSpace(doc.Text())
}
