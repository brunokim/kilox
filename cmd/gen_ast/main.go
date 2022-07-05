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
		log.Fatal(err)
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

	fmt.Fprintf(&b, `type %[1]s interface {
        accept(visitor %[2]sVisitor)
    }
    
    `, title, lower)
	fmt.Fprintf(&b, "type %sVisitor interface{\n", lower)
	for _, schema := range schemas {
		t := schemaType(schema.name, iName)
		fmt.Fprintf(&b, "\tvisit%s(%s %s)\n", t, lower, t)
	}
	fmt.Fprintf(&b, "}\n\n")
	for _, schema := range schemas {
		t := schemaType(schema.name, iName)
		fmt.Fprintf(&b, "type %s struct{\n", t)
		for _, field := range schema.fields {
			fmt.Fprintf(&b, "\t%s %s\n", field.name, field.type_)
		}
		fmt.Fprintf(&b, "}\n\n")
	}
	for _, schema := range schemas {
		fmt.Fprintf(&b, `func (%[1]s %[2]s) accept(v %[1]sVisitor) {
            v.visit%[2]s(%[1]s)
        }

        `, lower, schemaType(schema.name, iName))
	}
	return b.String()
}

func schemaType(sName, iName string) string {
	return sName + strings.Title(iName)
}

// ----

func interfaceName(filename string) string {
	filenameRE := regexp.MustCompile(`^.*/([^/]*).spec$`)
	groups := filenameRE.FindStringSubmatch(filename)
	return groups[1]
}

func structSchema(line string) schema {
	lineRE := regexp.MustCompile(`^(.*)\((.*)\)$`)
	parts := lineRE.FindStringSubmatch(line)
	return schema{
		name:   parts[1],
		fields: structFields(parts[2]),
	}
}

func structFields(list string) []field {
	params := strings.Split(list, ",")
	fields := make([]field, len(params))
	for i, param := range params {
		fields[i] = structField(strings.TrimSpace(param))
	}
	return fields
}

func structField(decl string) field {
	parts := strings.Split(decl, ":")
	return field{
		name:  strings.TrimSpace(parts[0]),
		type_: strings.TrimSpace(parts[1]),
	}
}
