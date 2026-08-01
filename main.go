package main

import (
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/http"
	"os"

	"github.com/RichardLeng88/localops/internal/project"
)

var pageTemplate = template.Must(template.New("inspection").Parse(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>LocalOps</title></head>
<body>
<main>
<h1>LocalOps project inspection</h1>
<form method="post" action="/">
<label for="project-path">Absolute project path</label>
<input id="project-path" name="project_path" required>
<button type="submit">Inspect</button>
</form>
{{if .Error}}<p role="alert">{{.Error}}</p>{{end}}
{{if .Result}}<dl>
<dt>Cleaned project path</dt><dd>{{.Result.ProjectPath}}</dd>
<dt>Git marker path</dt><dd>{{.Result.GitMarkerPath}}</dd>
<dt>Git marker type</dt><dd>{{.Result.GitMarkerType}}</dd>
</dl>{{end}}
</main>
</body>
</html>`))

type pageData struct {
	Error  string
	Result *project.Inspection
}

func main() {
	listener, err := listen()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer listener.Close()

	address := listener.Addr().String()
	fmt.Printf("http://%s\n", address)
	if err := http.Serve(listener, newHandler(address)); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func listen() (net.Listener, error) {
	return net.Listen("tcp4", "127.0.0.1:0")
}

func newHandler(listenerAddress string) http.Handler {
	return newHandlerWithInspect(listenerAddress, project.Inspect)
}

func newHandlerWithInspect(listenerAddress string, inspect func(string) (project.Inspection, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodGet {
			renderPage(w, http.StatusOK, pageData{})
			return
		}
		if r.Method != http.MethodPost {
			renderPage(w, http.StatusMethodNotAllowed, pageData{Error: "Inspection requests must use the form."})
			return
		}
		if r.Host != listenerAddress || r.Header.Get("Origin") != "http://"+listenerAddress {
			renderPage(w, http.StatusForbidden, pageData{Error: "Inspection request was rejected."})
			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/x-www-form-urlencoded" {
			renderPage(w, http.StatusBadRequest, pageData{Error: "Inspection requests must use the form."})
			return
		}
		if err := r.ParseForm(); err != nil {
			renderPage(w, http.StatusBadRequest, pageData{Error: "Enter an absolute project path."})
			return
		}
		paths := r.PostForm["project_path"]
		if len(paths) != 1 {
			renderPage(w, http.StatusBadRequest, pageData{Error: "Enter one absolute project path."})
			return
		}

		inspection, err := inspect(paths[0])
		if err != nil {
			renderPage(w, http.StatusBadRequest, pageData{Error: "The selected path is not an inspectable Git project."})
			return
		}
		renderPage(w, http.StatusOK, pageData{Result: &inspection})
	})
}

func renderPage(w http.ResponseWriter, status int, data pageData) {
	w.WriteHeader(status)
	_ = pageTemplate.Execute(w, data)
}
