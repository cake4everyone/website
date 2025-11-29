package webserver

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
)

var (
	rootPath string
)

const (
	htmlRoot     = "html"
	templateRoot = "template"
)

func Start(path string) {
	rootPath = path
	var (
		host = ""
		port = "8080"
		addr = net.JoinHostPort(host, port)
	)
	go func() {
		if err := http.ListenAndServe(addr, initHandler()); err != nil {
			panic(err)
		}
	}()
	time.Sleep(500 * time.Millisecond)
	log.Printf("Server started on %s", addr)
}

func initHandler() http.Handler {
	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(handle404)

	r.HandleFunc("/login", handleLoginPage).Methods(http.MethodGet)
	r.HandleFunc("/login", handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/account", handleAccountPage).Methods(http.MethodGet)
	r.HandleFunc("/account", handleAccount).Methods(http.MethodPatch)
	r.HandleFunc("/admin", handleAdminPage).Methods(http.MethodGet)
	r.HandleFunc("/admin", handleAdmin).Methods(http.MethodPatch)

	r.HandleFunc("/api/name/{uuid}", handleAPINameLookup).Methods(http.MethodGet)

	fileServer := http.FileServer(http.Dir(path.Join(rootPath, htmlRoot)))
	r.PathPrefix("/assets").Handler(http.StripPrefix("/assets", fileServer))

	return r
}

func handle404(w http.ResponseWriter, r *http.Request) {
	log.Printf("404: [%s] %s %s", r.Header.Get("X-Forwarded-For"), r.Method, r.RequestURI)
	// w.WriteHeader(http.StatusNotFound)

	ServeFile(w, r, "404")
}

func ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	if path.Ext(name) == "" {
		name += ".html"
	}
	http.ServeFile(w, r, path.Join(rootPath, htmlRoot, name))
}

func ServeTemplate(w http.ResponseWriter, r *http.Request, v any, name ...string) {
	for i, n := range name {
		if path.Ext(n) == "" {
			name[i] += ".html"
		}
	}
	tmpl, err := ParseTemplate(name...)
	if err != nil {
		log.Printf("Could not parse %s template: %+v", name, err)
		http.Error(w, fmt.Sprintf("Failed to load %s page", name), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, v)
	if err != nil {
		log.Printf("Could not execute %s template: %+v", name, err)
		http.Error(w, fmt.Sprintf("Failed to load %s page", name), http.StatusInternalServerError)
		return
	}
}
