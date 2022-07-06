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
)

var (
	spec = flag.String("spec", "", "Spec file to read")
	pkg  = flag.String("pkg", "lox", "Package where this code belongs")
	dest = flag.String("dest", "", "Destination file. If not given, default to gen_%s.go, where %s is the spec name")
)

type field struct {
	name  string
	type_ string
}

type schema struct {
	name   string
	fields []field
	isPtr  bool
}

func main() {
	flag.Parse()
	bs, err := ioutil.ReadFile(*spec)
	if err != nil {
		log.Fatal(err)
	}
	iName := interfaceName(*spec)
	if *dest == "" {
		*dest = fmt.Sprintf("gen_%s.go", iName)
	}
	lines := strings.Split(string(bs), "\n")
	var schemas []schema
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		schemas = append(schemas, structSchema(line))
	}
	src := writeSource(iName, schemas)
	bs, err = format.Source([]byte(src))
	if err != nil {
		log.Print(src)
		log.Fatal("gofmt:", err)
	}
	err = ioutil.WriteFile(*dest, bs, 0664)
	if err != nil {
		log.Fatal(err)
	}
}

func writeSource(iName string, schemas []schema) string {
	lower, title := strings.ToLower(iName), strings.Title(iName)
	var b strings.Builder
	fmt.Fprintf(&b, "// Generated file, do not modify\n")
	fmt.Fprintf(&b, "// Invocation: gen_ast %s\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintf(&b, "package %s\n\n", *pkg)

	// Interface declaration.
	fmt.Fprintf(&b, `type %[1]s interface {
        accept(v %[2]sVisitor)
    }
    
    `, title, lower)

	// Visitor declaration.
	fmt.Fprintf(&b, "type %sVisitor interface{\n", lower)
	for _, schema := range schemas {
		name := schemaName(schema, iName)
		t := schemaType(schema, iName)
		fmt.Fprintf(&b, "\tvisit%s(%c %s)\n", name, lower[0], t)
	}
	fmt.Fprintf(&b, "}\n\n")

	// Interface instances declaration.
	for _, schema := range schemas {
		name := schemaName(schema, iName)
		fmt.Fprintf(&b, "type %s struct{\n", name)
		for _, field := range schema.fields {
			fmt.Fprintf(&b, "\t%s %s\n", field.name, field.type_)
		}
		fmt.Fprintf(&b, "}\n\n")
	}

	// Instances implementation of accept.
	for _, schema := range schemas {
		name := schemaName(schema, iName)
		t := schemaType(schema, iName)
		fmt.Fprintf(&b, `func (%[1]c %[2]s) accept(v %[3]sVisitor) {
            v.visit%[4]s(%[1]c)
        }

        `, lower[0], t, lower, name)
	}
	return b.String()
}

func schemaName(s schema, iName string) string {
	return s.name + strings.Title(iName)
}

func schemaType(s schema, iName string) string {
	name := schemaName(s, iName)
	if s.isPtr {
		return "*" + name
	}
	return name
}

// ----

func interfaceName(filename string) string {
	filenameRE := regexp.MustCompile(`^.*/([^/]*).spec$`)
	groups := filenameRE.FindStringSubmatch(filename)
	if groups == nil {
		panic(fmt.Sprintf("spec name %q doesn't end with '.spec'", filename))
	}
	return groups[1]
}

func structSchema(line string) schema {
	lineRE := regexp.MustCompile(`^\s*(\*)?(.*)\((.*)\)$`)
	parts := lineRE.FindStringSubmatch(line)
	if parts == nil {
		panic(fmt.Sprintf("line %q doesn't match pattern 'Struct(Field1: Type1, Field2: Type2)'", line))
	}
	return schema{
		name:   parts[2],
		fields: structFields(parts[3]),
		isPtr:  parts[1] == "*",
	}
}

func structFields(list string) []field {
	params := strings.Split(list, ",")
	var fields []field
	for _, param := range params {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}
		fields = append(fields, structField(param))
	}
	return fields
}

func structField(decl string) field {
	parts := strings.Split(decl, ":")
	if len(parts) != 2 {
		panic(fmt.Sprintf("decl %q doesn't match with 'Name: Type' pattern", decl))
	}
	return field{
		name:  strings.TrimSpace(parts[0]),
		type_: strings.TrimSpace(parts[1]),
	}
}
