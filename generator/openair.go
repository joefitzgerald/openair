package generator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Date is a date
const Date string = "Date"

// Address is an address
const Address string = "Address"

// Config is OpenAir configuration
type Config struct {
	Scheme    string `default:"https"`
	Domain    string `default:"sandbox.openair.com"`
	Key       string `required:"true"`
	Namespace string `default:"default"`
	Company   string `required:"true"`
	User      string `required:"true"`
	Password  string `required:"true"`
}

type generator struct {
	c            Config
	objectNames  string
	dir          string
	pkg          string
	outputPrefix string
	outputSuffix string
}

// OpenAirGenerator generates an API client for the OpenAir XML API
type OpenAirGenerator interface {
	GenerateCommonFile()
	GenerateModelFiles()
}

// New creates a generator
func New(c Config, objectNames string, dir string, outputPrefix string, outputSuffix string) OpenAirGenerator {
	g := &generator{c: c, objectNames: objectNames, dir: dir, outputPrefix: outputPrefix, outputSuffix: outputSuffix}
	pkg, err := GetPackageName(dir, g.outputPrefix, g.outputSuffix+".go")
	if err != nil {
		log.Fatal(err)
	}
	g.pkg = pkg
	return g
}

type field struct {
	FieldName string
	RawName   string
	FieldType string
}

func fetchFromOpenAir(datatype string) ([]byte, error) {
	var c Config
	err := envconfig.Process("openair", &c)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s://%s/api.pl", c.Scheme, c.Domain)
	tmpl := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
  <request API_version="1.0" client_ver="1.1"
  namespace="default" key="%s">
    <Auth>
      <Login>
        <company>%s</company>
        <user>%s</user>
        <password>%s</password>
      </Login>
    </Auth>
    <Read type="%s" method="all" limit="1" enable_custom="1" include_nondeleted="1" deleted="1" />
  </request>`

	payload := strings.NewReader(fmt.Sprintf(tmpl, c.Key, c.Company, c.User, c.Password, datatype))
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/xml")
	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(res.Body)
}

func buildFields(obj string) []field {
	var r Response
	var fields []field
	body, err := fetchFromOpenAir(obj)
	if err != nil {
		log.Fatal(err)
	}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range r.Read.Entity.Element {
		clean := cleanname(e.XMLName.Local)
		t := "string"
		if len(e.Element) > 0 && e.Element[0].XMLName.Local == Date {
			t = Date
		}
		if len(e.Element) > 0 && e.Element[0].XMLName.Local == Address {
			t = Address
		}
		fields = append(fields, field{FieldName: clean, RawName: e.XMLName.Local, FieldType: t})
	}

	for _, f := range fields {
		count := 1
		for i, f2 := range fields {
			if f.FieldName == f2.FieldName && f.RawName != f2.RawName {
				f3 := &fields[i]
				if strings.Contains(f3.RawName, "_") {
					f3.FieldName = f3.FieldName + strconv.Itoa(count)
				}
				count = count + 1
			}
		}
	}

	return fields
}

func (g *generator) GenerateModelFiles() {
	datatypes := strings.Split(g.objectNames, ",")
	for _, datatype := range datatypes {
		name := cleanname(datatype)
		fields := buildFields(datatype)
		var context = struct {
			PackageName string
			TypeName    string
			Fields      []field
		}{
			PackageName: g.pkg,
			TypeName:    name,
			Fields:      fields,
		}

		var buf bytes.Buffer
		if err := generatedTmpl.Execute(&buf, context); err != nil {
			log.Fatalf("generating code: %v", err)
		}

		src, err := format.Source(buf.Bytes())
		if err != nil {
			log.Printf("warning: internal error: invalid Go generated: %s", err)
			log.Printf("warning: compile the package to analyze the error")
			src = buf.Bytes()
		}

		output := strings.ToLower(g.outputPrefix + context.TypeName + g.outputSuffix + ".go")
		outputPath := filepath.Join(g.dir, output)
		if err := ioutil.WriteFile(outputPath, src, 0644); err != nil {
			log.Fatalf("writing output: %s", err)
		}
	}
}

func (g *generator) GenerateCommonFile() {
	var buf bytes.Buffer
	datatypes := strings.Split(g.objectNames, ",")
	var context = struct {
		PackageName string
		Types       []string
	}{
		PackageName: g.pkg,
		Types:       datatypes,
	}

	if err := commonTmpl.Execute(&buf, context); err != nil {
		log.Fatalf("generating code: %v", err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		src = buf.Bytes()
	}

	output := strings.ToLower(g.outputPrefix + "common" + g.outputSuffix + ".go")
	outputPath := filepath.Join(g.dir, output)
	if err := ioutil.WriteFile(outputPath, src, 0644); err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// Element contains an element
type Element struct {
	XMLName xml.Name
	Element []Element `xml:",any"`
	Value   string    `xml:",chardata"`
}

// Response is a container for Auth and Read requests
type Response struct {
	XMLName xml.Name `xml:"response"`
	Auth    Auth     `xml:"Auth,omitempty"`
	Read    Read     `xml:"Read,omitempty"`
}

// Auth includes status information about the authorization of a request
type Auth struct {
	Status string `xml:"status,attr"`
}

// Read is a container for OpenAir entities
type Read struct {
	XMLName xml.Name `xml:"Read"`
	Status  string   `xml:"status,attr"`
	Entity  Element  `xml:",any"`
}

// GetPackageName finds the package name for the given directory
func GetPackageName(directory, skipPrefix, skipSuffix string) (string, error) {
	pkgDir, err := build.Default.ImportDir(directory, 0)
	if err != nil {
		return "", fmt.Errorf("cannot process directory %s: %s", directory, err)
	}

	var files []*ast.File
	fs := token.NewFileSet()
	for _, name := range pkgDir.GoFiles {
		if !strings.HasSuffix(name, ".go") ||
			(skipSuffix != "" && strings.HasPrefix(name, skipPrefix) &&
				strings.HasSuffix(name, skipSuffix)) {
			continue
		}
		if directory != "." {
			name = filepath.Join(directory, name)
		}
		f, err := parser.ParseFile(fs, name, nil, 0)
		if err != nil {
			return "", fmt.Errorf("parsing file %v: %v", name, err)
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("%s: no buildable Go files", directory)
	}

	// type-check the package
	defs := make(map[*ast.Ident]types.Object)
	config := types.Config{FakeImportC: true, Importer: importer.Default()}
	info := &types.Info{Defs: defs}
	if _, err := config.Check(directory, fs, files, info); err != nil {
		return "", fmt.Errorf("type-checking package: %v", err)
	}

	return files[0].Name.Name, nil
}
