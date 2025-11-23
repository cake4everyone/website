package webserver

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"website/database"

	"github.com/google/uuid"
)

var functionMap = make(template.FuncMap)

func init() {
	functionMap["GetMCName"] = GetMCName
	functionMap["IsAdmin"] = func(e database.WhitelistEntry) bool { return e.IsAdmin() }
}

func ParseTemplate(file string) (tmpl *template.Template, err error) {
	return template.New(path.Base(file)).
		Funcs(GetFunctionMap()).
		ParseFiles(
			path.Join(rootPath, templateRoot, file),
			path.Join(rootPath, templateRoot, "common.html"),
		)
}

func GetFunctionMap() template.FuncMap {
	return functionMap
}

func GetMCName(uuid uuid.UUID) string {
	resp, err := http.Get(fmt.Sprintf("https://api.minecraftservices.com/minecraft/profile/lookup/%s", uuid))
	if err != nil {
		log.Printf("Failed to resolve MC name: %v", err)
		return "ERROR:?"
	}
	if resp.StatusCode != 200 {
		return "ERROR:" + resp.Status
	}

	var data struct {
		Name string `json:"name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Failed to decode MC name response: %v", err)
		return "ERROR:?"
	}
	return data.Name
}
