package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/cilium/ebpf"
)

// Stores all routes to display them at "/"
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

// EbpfMapPrintHandler is a handler that can be used to print the content of an eBPF map.
type EbpfMapPrintHandler struct {
	ebpfMap   *ebpf.Map
	formatter func(*ebpf.Map) (string, error)
}

func (eh EbpfMapPrintHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
