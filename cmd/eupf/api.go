package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cilium/ebpf"
)

var webAddr = flag.String("waddr", ":8080", "Address to bind web server to")

// Stores all routes to display them at /
type RootHandler struct {
	routes []string
}

func (r *RootHandler) AddRoute(route string) {
	r.routes = append(r.routes, route)
}

func (rh RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `<!DOCTYPE html><html lang="en-US"><body>`)
	for _, route := range rh.routes {
		io.WriteString(w, fmt.Sprintf(`<a href="/%[1]s">%[1]s</a></br>`, route))
	}
	io.WriteString(w, `</body></html>`)
}

// Handles member eBPF map separately with supplied formatter function.
type EbpfMapHandler struct {
	ebpfMap   *ebpf.Map
	formatter func(*ebpf.Map) (string, error)
}

func (eh EbpfMapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		mapContent, err := FormatMapContents(eh.ebpfMap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, mapContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Only GET method is implemented")
	}
}

// Use same function for provided maps
func StartAPI(ebpfMaps ...*ebpf.Map) {
	mux := http.NewServeMux()
	rootHandler := RootHandler{}
	for _, ebpfMap := range ebpfMaps {
		info, err := ebpfMap.Info()
		if err != nil {
			panic(err)
		}
		log.Printf("Creating web api route for %s", info.Name)
		mux.Handle(fmt.Sprintf("/%s", info.Name), EbpfMapHandler{ebpfMap: ebpfMap, formatter: FormatMapContents})
		rootHandler.AddRoute(info.Name)
	}
	mux.Handle("/", rootHandler)
	log.Printf("Web server started on address: %s", *webAddr)
	http.ListenAndServe(*webAddr, mux)
}

// Or someting like builder pattern
type ApiBuilder struct {
	mux         *http.ServeMux
	rootHandler RootHandler
}

func NewApiBuilder() *ApiBuilder {
	return &ApiBuilder{
		mux:         http.NewServeMux(),
		rootHandler: RootHandler{},
	}
}

func (b *ApiBuilder) AddMap(route string, ebpfMap *ebpf.Map, formatter func(*ebpf.Map) (string, error)) {	
	info, err := ebpfMap.Info()
	if err != nil {
		panic(err)
	}
	b.rootHandler.AddRoute(route)
	b.mux.Handle(fmt.Sprintf("/%s", route), EbpfMapHandler{ebpfMap: ebpfMap, formatter: formatter})
	log.Printf("Added route %s to map %s", route, info.Name)
}

func (b *ApiBuilder) StartAPI() {
	b.mux.Handle("/", b.rootHandler)
	log.Printf("Web server started on address: %s", *webAddr)
	http.ListenAndServe(*webAddr, b.mux)
}
