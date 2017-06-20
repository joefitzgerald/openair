package generator

import (
	"strings"
	"text/template"
)

func cleanname(name string) string {
	if strings.Contains(name, "_") {
		newName := ""
		splitName := strings.Split(name, "_")
		for _, item := range splitName {
			if len(item) == 0 {
				continue
			}
			if len(item) == 1 {
				newName += strings.ToUpper(item)
				continue
			}
			if len(item) == 2 {
				if strings.ToUpper(item) == "ID" {
					newName += strings.ToUpper(item)
					continue
				}
			}
			if (item) == "approvalprocess" {
				newName += "ApprovalProcess"
				continue
			}
			newName += strings.ToUpper(string(item[0])) + string(item[1:])
		}
		name = newName
	}

	if strings.HasSuffix(name, "Id") {
		name = strings.TrimSuffix(name, "Id") + "ID"
	}
	if strings.HasSuffix(name, "id") {
		name = strings.TrimSuffix(name, "id") + "ID"
	}
	if strings.HasSuffix(name, "Url") {
		name = strings.TrimSuffix(name, "Url") + "URL"
	}
	if strings.HasSuffix(name, "url") {
		name = strings.TrimSuffix(name, "url") + "URL"
	}
	if strings.HasSuffix(name, "Api") {
		name = strings.TrimSuffix(name, "Api") + "API"
	}
	if strings.HasSuffix(name, "api") {
		name = strings.TrimSuffix(name, "api") + "API"
	}
	if strings.HasPrefix(name, "Api") {
		name = "API" + strings.TrimPrefix(name, "Api")
	}
	if strings.HasPrefix(name, "Url") {
		name = "URL" + strings.TrimPrefix(name, "Url")
	}

	return strings.Title(name)
}

func tag(tagname string, t string) string {
	xmlname := tagname
	jsonname := tagname
	if strings.ToLower(t) == strings.ToLower(Address) {
		xmlname = tagname + ">" + Address
	}
	if strings.ToLower(t) == strings.ToLower(Date) {
		xmlname = tagname + ">" + Date
	}
	return "`xml:\"" + xmlname + ",omitempty\" json:\"" + jsonname + ",omitempty\"`"
}

func xmltag(tagName string) string {
	return "`xml:\"" + tagName + "\"`"
}

func xmlrawtag(tagName string) string {
	return "`xml:\"" + tagName + ",omitempty\"`"
}

func cleannamelower(name string) string {
	return strings.ToLower(cleanname(name))
}

func valueforkey(key string, m map[string]string) string {
	return m[key]
}

func backtick() string {
	return "`"
}

var generatedTmpl = template.Must(template.New("generated").Funcs(template.FuncMap{
	"tag":            tag,
	"backtick":       backtick,
	"xmltag":         xmltag,
	"xmlrawtag":      xmlrawtag,
	"cleanname":      cleanname,
	"cleannamelower": cleannamelower,
	"tolower":        strings.ToLower,
}).Parse(`
// Code generated by openair; DO NOT EDIT.

package {{.PackageName}}

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"time"
	"net/http"
	"strings"
)

// {{cleanname .TypeName}} is the {{.TypeName}} OpenAir XML Datatype
type {{cleanname .TypeName}} struct {
	{{range .Fields}}{{cleanname .FieldName}} {{.FieldType}} {{tag .RawName .FieldType}}
	{{end}}
}

// {{cleanname .TypeName}}Response is a container for Auth and Read requests
type {{cleanname .TypeName}}Response struct {
	XMLName xml.Name     {{xmltag "response"}}
	Auth    Auth         {{xmltag "Auth,omitempty"}}
	Read    {{cleanname .TypeName}}Read {{xmltag "Read,omitempty"}}
}

// {{cleanname .TypeName}}Read is a container for {{cleanname .TypeName}}
type {{cleanname .TypeName}}Read struct {
	Status    string     {{xmltag "status,attr"}}
	{{cleanname .TypeName}}s []{{cleanname .TypeName}} {{xmlrawtag .TypeName}}
}

type {{cleannamelower .TypeName}} struct {
	config *Config
}

func (o *{{cleannamelower .TypeName}}) list(ctx context.Context, limit int, offset int, modifiedSince *time.Time) ([]{{cleanname .TypeName}}, error) {
	url := fmt.Sprintf("%s://%s/api.pl", o.config.Scheme, o.config.Domain)

	filterAttributes := ""
	filterBody := ""

	if modifiedSince != nil {
		filterAttributes = "filter=\"newer-than\" field=\"date\""
		filterBody = fmt.Sprintf({{backtick}}<Date>
			<year>%d</year>
			<month>%d</month>
			<day>%d</day>
		</Date>{{backtick}}, modifiedSince.Year(), modifiedSince.Month(), modifiedSince.Day())
	}

	tmpl := {{backtick}}<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<request API_version="1.0" client_ver="1.1" namespace="%s" key="%s">
			<Auth>
				<Login>
					<company>%s</company>
					<user>%s</user>
					<password>%s</password>
				</Login>
			</Auth>
			<Read type="{{.TypeName}}" method="all" limit="%d,%d" enable_custom="1" include_nondeleted="1" deleted="1" %s>%s</Read>
		</request>{{backtick}}

	payload := strings.NewReader(fmt.Sprintf(tmpl, o.config.Namespace, o.config.Key, o.config.Company, o.config.User, o.config.Password, offset, limit, filterAttributes, filterBody))

	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "application/xml")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var r {{cleanname .TypeName}}Response
	if err := xml.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.Auth.Status != "0" {
		return nil, errors.New("unauthorized")
	}
	return r.Read.{{cleanname .TypeName}}s, nil
}

func (o *{{cleannamelower .TypeName}}) listWithRetry(ctx context.Context, limit int, offset int, modifiedSince *time.Time) ([]{{cleanname .TypeName}}, error) {
	wait := time.Duration(o.config.RetryDelay) * time.Millisecond
	attempt := 1
	batch, err := o.list(ctx, limit, offset, modifiedSince)
	if err != nil && err.Error() == "unauthorized" {
		return nil, err
	}
	for err != nil {
		if attempt == 8 {
			return nil, err
		}
		time.Sleep(wait)
		wait *= 2
		attempt += 1
		batch, err = o.list(ctx, limit, offset, modifiedSince)
	}
	return batch, nil
}

func (o *{{cleannamelower .TypeName}}) ListAsync(ctx context.Context, modifiedSince *time.Time) (<-chan []{{cleanname .TypeName}}, <-chan error) {
	result := make(chan []{{cleanname .TypeName}})
	errs := make(chan error, 1)

	limit := 1000

	go func() {
		offset := 0
		for {
			batch, err := o.listWithRetry(ctx, limit, offset, modifiedSince)
			if err != nil {
				errs <- err
				break
			}
			result <- batch
			if len(batch) < limit {
				break
			}
			offset += limit
		}
		close(result)
		close(errs)
	}()

	return result, errs
}

`))

var commonTmpl = template.Must(template.New("common").Funcs(template.FuncMap{
	"tag":            tag,
	"xmltag":         xmltag,
	"cleanname":      cleanname,
	"cleannamelower": cleannamelower,
	"tolower":        strings.ToLower,
	"valueforkey":    valueforkey,
	"backtick":       backtick,
}).Parse(`
// Code generated by openair; DO NOT EDIT.

package {{.PackageName}}

import "github.com/kelseyhightower/envconfig"

// API is an OpenAir XML API client.
type API struct {
	config *Config
	{{range $idx, $value := .Types}}{{cleanname $value}} *{{cleannamelower $value}}
{{end}}
}

// Config is OpenAir configuration
type Config struct {
	Scheme     string {{backtick}}default:"https"{{backtick}}
	Domain     string {{backtick}}default:"sandbox.openair.com"{{backtick}}
	Key        string {{backtick}}required:"true"{{backtick}}
	Namespace  string {{backtick}}default:"default"{{backtick}}
	Company    string {{backtick}}required:"true"{{backtick}}
	User       string {{backtick}}required:"true"{{backtick}}
	Password   string {{backtick}}required:"true"{{backtick}}
	RetryDelay int    {{backtick}}default:"100"{{backtick}}
}

// New creates a new OpenAir API, making use of the environment to generate a Config
func New() (*API, error) {
	var c Config
	err := envconfig.Process("openair", &c)
	if err != nil {
		return nil, err
	}

	return NewWithConfig(&c), nil
}

// NewWithConfig creates a new OpenAir API with the provided Config
func NewWithConfig(c *Config) *API {
	api := &API{
	config: c,
	{{range $idx, $value := .Types}}{{cleanname $value}}: &{{cleannamelower $value}}{ config: c, },
	{{end}}
	}

	return api
}

// Auth includes status information about the authorization of a request
type Auth struct {
	Status string {{xmltag "status,attr"}}
}

// Date is a date
type Date struct {
	Hour     string {{tag "hour" "string"}}
	Minute   string {{tag "minute" "string"}}
	Timezone string {{tag "timezone" "string"}}
	Second   string {{tag "second" "string"}}
	Month    string {{tag "month" "string"}}
	Day      string {{tag "day" "string"}}
	Year     string {{tag "year" "string"}}
}

// Address is an address
type Address struct {
	ID         string {{tag "id" "string"}}
	ContactID  string {{tag "contact_id" "string"}}
	Salutation string {{tag "salutation" "string"}}
	First      string {{tag "first" "string"}}
	Middle     string {{tag "middle" "string"}}
	Last       string {{tag "last" "string"}}
	Email      string {{tag "email" "string"}}
	Phone      string {{tag "phone" "string"}}
	Fax        string {{tag "fax" "string"}}
	Mobile     string {{tag "mobile" "string"}}
	Addr1      string {{tag "addr1" "string"}}
	Addr2      string {{tag "addr2" "string"}}
	Addr3      string {{tag "addr3" "string"}}
	Addr4      string {{tag "addr4" "string"}}
	City       string {{tag "city" "string"}}
	State      string {{tag "state" "string"}}
	Zip        string {{tag "zip" "string"}}
	Country    string {{tag "country" "string"}}
}
`))
