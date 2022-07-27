package main

import (
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type StringSet map[string]bool

type Data struct {
	Extensions StringSet
	Invocation string
	Name       string
	VarName    string
	Schemas    []Schema
}

type Schema struct {
	Name   string
	IsPtr  bool
	Fields []Field
}

type Field struct {
	Name string
	Type string
}

func schemaName(d Data, s Schema) string {
	return s.Name + d.Name
}

func schemaType(d Data, s Schema) string {
	name := schemaName(d, s)
	if s.IsPtr {
		return "*" + name
	}
	return name
}

const fileTemplate = `
{{- block "header" . -}}
// Generated file, do not modify
// Invocation: gen_ast {{.Invocation}}
package lox
{{- end}}

{{block "interface declaration" . -}}
type {{.Name}} interface {
	Accept(v {{.Name | lower}}Visitor)
    {{if .Extensions.typename}}TypeName() string{{end}}
}
{{- end}}

{{block "visitor declaration" . -}}
type {{.Name|lower}}Visitor interface{
	{{range .Schemas -}}
		Visit{{schemaName $ .}}({{$.VarName}} {{schemaType $ .}})
	{{end -}}
}
{{- end}}

{{block "struct declaration" . -}}
{{range .Schemas -}}
    type {{schemaName $ .}} struct {
        {{range .Fields -}}
            {{.Name}} {{.Type}}
        {{end -}}
    }

{{end -}}
{{- end}}

{{block "interface implementation" . -}}
{{range .Schemas -}}
    func ({{$.VarName}} {{schemaType $ .}}) Accept(v {{$.Name | lower}}Visitor) {
        v.Visit{{schemaName $ .}}({{$.VarName}})
    }

{{end -}}
{{- end}}

{{block "typename" . -}}
{{if .Extensions.typename}}
    {{range .Schemas -}}
        func ({{schemaType $ .}}) TypeName() string { return "{{.Name | lower}}"; }
    {{end -}}
{{end}}
{{- end}}
`

var tmpl = template.Must(
	template.New("").Funcs(map[string]any{
		"schemaType": schemaType,
		"schemaName": schemaName,
		"lower":      strings.ToLower,
	}).Parse(fileTemplate))

var (
	spec       = flag.String("spec", "", "Spec file to read")
	pkg        = flag.String("pkg", "lox", "Package where this code belongs")
	dest       = flag.String("dest", "", "Destination file. If not given, default to gen_%s.go, where %s is the spec name")
	extensions = flag.String("extensions", "", "Comma-separated list of extensions to apply")
)

func main() {
	flag.Parse()
	iName := interfaceName(*spec)
	if *dest == "" {
		*dest = fmt.Sprintf("gen_%s.go", iName)
	}
	bs, err := ioutil.ReadFile(*spec)
	if err != nil {
		log.Fatal(err)
	}
	exts := parseExtensions(*extensions)
	data := &Data{
		Extensions: exts,
		Invocation: strings.Join(os.Args[1:], " "),
		Name:       strings.Title(iName),
		VarName:    iName[:1],
		Schemas:    parseSchemas(string(bs)),
	}
	if data.VarName == "v" {
		// Conflicts with visitor variable, use another.
		data.VarName = "x"
	}
	var b strings.Builder
	err = tmpl.Execute(&b, data)
	if err != nil {
		log.Fatal("template:", err)
	}
	bs, err = format.Source([]byte(b.String()))
	if err != nil {
		log.Print(b.String())
		log.Fatal("gofmt:", err)
	}
	err = ioutil.WriteFile(*dest, bs, 0664)
	if err != nil {
		log.Fatal(err)
	}
}

// ----

func interfaceName(filename string) string {
	filenameRE := regexp.MustCompile(`^.*/([^/]*).spec$`)
	groups := filenameRE.FindStringSubmatch(filename)
	if groups == nil {
		panic(fmt.Sprintf("spec name %q doesn't end with '.spec'", filename))
	}
	return strings.ToLower(groups[1])
}

func parseExtensions(text string) StringSet {
	exts := make(StringSet)
	extList := strings.Split(text, ",")
	for _, ext := range extList {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext == "" {
			continue
		}
		exts[ext] = true
	}
	return exts
}

func parseSchemas(text string) []Schema {
	var schemas []Schema
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		commentRE := regexp.MustCompile(`^(.*?)(//.*)?$`)
		groups := commentRE.FindStringSubmatch(line)
		line = strings.TrimSpace(groups[1])
		if line == "" {
			continue
		}
		schemas = append(schemas, parseSchema(line))
	}
	return schemas
}

func parseSchema(line string) Schema {
	lineRE := regexp.MustCompile(`^\s*(\*)?(.*)\((.*)\)$`)
	parts := lineRE.FindStringSubmatch(line)
	if parts == nil {
		panic(fmt.Sprintf("line %q doesn't match pattern 'Struct(Field1: Type1, Field2: Type2)'", line))
	}
	return Schema{
		Name:   parts[2],
		IsPtr:  parts[1] == "*",
		Fields: parseFields(parts[3]),
	}
}

func parseFields(list string) []Field {
	params := strings.Split(list, ",")
	var fields []Field
	for _, param := range params {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}
		fields = append(fields, parseField(param))
	}
	return fields
}

func parseField(decl string) Field {
	parts := strings.Split(decl, ":")
	if len(parts) != 2 {
		panic(fmt.Sprintf("decl %q doesn't match with 'Name: Type' pattern", decl))
	}
	return Field{
		Name: strings.TrimSpace(parts[0]),
		Type: strings.TrimSpace(parts[1]),
	}
}
