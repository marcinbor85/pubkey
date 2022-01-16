package home

import (
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/marcinbor85/pubkey/config"
)

func Register(router *mux.Router) {

	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
    router.PathPrefix("/static/").Handler(s)

	router.HandleFunc("/", homePageEndpoint).Methods(http.MethodGet)
}

func homePageEndpoint(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles(config.Get("TEMPLATE_HOME_PAGE"))

	host := config.Get("HOST")

	type PageContext struct {
		HostAddress     string
	}

	context := PageContext{
		HostAddress:     host,
	}

	tmpl.Execute(w, context)
}
