package webserver

import (
	"html/template"
	"path"
	"website/database"
)

var functionMap = make(template.FuncMap)

func init() {
	functionMap["IsAdmin"] = func(e database.WhitelistEntry) bool { return e.IsAdmin() }
}

func ParseTemplate(file ...string) (tmpl *template.Template, err error) {
	file = append(file, "common.html")
	for i, f := range file {
		file[i] = path.Join(rootPath, templateRoot, f)
	}
	return template.New(path.Base(file[0])).
		Funcs(GetFunctionMap()).
		ParseFiles(file...)
}

func GetFunctionMap() template.FuncMap {
	return functionMap
}
