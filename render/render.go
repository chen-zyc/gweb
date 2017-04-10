package render

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type Render interface {
	Render(w http.ResponseWriter) error
}

type RenderFunc func(w http.ResponseWriter) error

func (rf RenderFunc) Render(w http.ResponseWriter) error {
	return rf(w)
}

func StringRender(format string, args ...interface{}) Render {
	return RenderFunc(func(w http.ResponseWriter) error {
		writeContentType(w, "text/plain; charset=utf-8")
		if len(args) > 0 {
			_, err := fmt.Fprintf(w, format, args...)
			return err
		}
		_, err := io.WriteString(w, format)
		return err
	})
}

func JSONRender(obj interface{}) Render {
	return RenderFunc(func(w http.ResponseWriter) error {
		writeContentType(w, "application/json; charset=utf-8")
		return json.NewEncoder(w).Encode(obj)
	})
}

func XMLRender(obj interface{}) Render {
	return RenderFunc(func(w http.ResponseWriter) error {
		writeContentType(w, "application/xml; charset=utf-8")
		return xml.NewEncoder(w).Encode(obj)
	})
}

func writeContentType(w http.ResponseWriter, ct string) {
	w.Header()["Content-Type"] = []string{ct}
}
