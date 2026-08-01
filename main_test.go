package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RichardLeng88/localops/internal/project"
)

const testListenerAddress = "127.0.0.1:41234"

func TestGetRendersInspectionForm(t *testing.T) {
	response := serve(t, newHandler(testListenerAddress), httptest.NewRequest(http.MethodGet, "http://"+testListenerAddress+"/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `<form method="post" action="/">`) || !strings.Contains(response.Body.String(), `name="project_path"`) {
		t.Fatalf("response does not contain the inspection form: %s", response.Body.String())
	}
}

func TestPostRendersInspectionFields(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "project&escaped")
	if err := os.Mkdir(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(projectPath, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	response := serve(t, newHandler(testListenerAddress), inspectionRequest(testListenerAddress, url.Values{"project_path": {projectPath}}))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, value := range []string{"Cleaned project path", "Git marker path", "Git marker type", strings.ReplaceAll(filepath.Clean(projectPath), "&", "&amp;"), strings.ReplaceAll(filepath.Join(projectPath, ".git"), "&", "&amp;"), "directory"} {
		if !strings.Contains(body, value) {
			t.Fatalf("response does not contain %q: %s", value, body)
		}
	}
	if strings.Contains(body, projectPath) {
		t.Fatalf("response contains unescaped path: %s", body)
	}
	if strings.Contains(body, "The selected path is not") {
		t.Fatalf("response contains an error: %s", body)
	}
	if strings.Count(body, "<dd>") != 3 {
		t.Fatalf("response contains result fields beyond the approved three: %s", body)
	}
}

func TestPostRejectsNonGitDirectoryEvenWithGitSibling(t *testing.T) {
	parent := t.TempDir()
	selected := filepath.Join(parent, "selected")
	sibling := filepath.Join(parent, "sibling")
	if err := os.Mkdir(selected, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(sibling, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	response := serve(t, newHandler(testListenerAddress), inspectionRequest(testListenerAddress, url.Values{"project_path": {selected}}))
	assertInspectionError(t, response)
}

func TestPostErrorsDoNotRetainPath(t *testing.T) {
	handler := newHandler(testListenerAddress)
	for name, values := range map[string]url.Values{
		"missing": {},
		"invalid": {"project_path": {filepath.Join(t.TempDir(), "missing")}},
	} {
		t.Run(name, func(t *testing.T) {
			response := serve(t, handler, inspectionRequest(testListenerAddress, values))
			assertInspectionError(t, response)
		})
	}

	response := serve(t, handler, httptest.NewRequest(http.MethodGet, "http://"+testListenerAddress+"/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "missing") || strings.Contains(response.Body.String(), "inspectable Git project") {
		t.Fatalf("GET retained an earlier submission: %s", response.Body.String())
	}
}

func TestInspectionRejectsUntrustedRequestsBeforeInspecting(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "project")
	if err := os.Mkdir(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(projectPath, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	nonFormRequest := originRequest(http.MethodPost, testListenerAddress, "http://"+testListenerAddress, nil)
	nonFormRequest.Body = io.NopCloser(strings.NewReader(url.Values{"project_path": {projectPath}}.Encode()))
	nonFormRequest.Header.Set("Content-Type", "text/plain")
	wrongHostRequest := inspectionRequest(testListenerAddress, url.Values{"project_path": {projectPath}})
	wrongHostRequest.Host = "127.0.0.1:49999"
	handler := newHandlerWithInspect(testListenerAddress, func(string) (project.Inspection, error) {
		t.Fatal("rejected request reached inspection")
		return project.Inspection{}, nil
	})
	for name, request := range map[string]*http.Request{
		"wrong Host":          wrongHostRequest,
		"missing Origin":      formRequest(http.MethodPost, testListenerAddress, url.Values{"project_path": {projectPath}}),
		"cross-origin":        originRequest(http.MethodPost, testListenerAddress, "http://127.0.0.1:49999", url.Values{"project_path": {projectPath}}),
		"non-form":            nonFormRequest,
		"non-POST inspection": originRequest(http.MethodPut, testListenerAddress, "http://"+testListenerAddress, url.Values{"project_path": {projectPath}}),
	} {
		t.Run(name, func(t *testing.T) {
			response := serve(t, handler, request)
			assertInspectionError(t, response)
		})
	}
}

func TestListenerUsesAssignedTCP4LoopbackAddress(t *testing.T) {
	listener, err := listen()
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("listener address type = %T", listener.Addr())
	}
	if address.Port == 0 {
		t.Fatal("listener did not receive an OS-assigned port")
	}
	if address.IP.IsUnspecified() || !address.IP.Equal(net.IPv4(127, 0, 0, 1)) || address.IP.To4() == nil {
		t.Fatalf("listener address = %s", address)
	}
}

func TestHandlerProducesNoRequestLog(t *testing.T) {
	projectPath := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectPath, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	output := captureOutput(t, func() {
		response := serve(t, newHandler(testListenerAddress), inspectionRequest(testListenerAddress, url.Values{"project_path": {projectPath}}))
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d", response.Code)
		}
	})
	if output != "" {
		t.Fatalf("request output = %q", output)
	}
}

func inspectionRequest(listenerAddress string, values url.Values) *http.Request {
	request := formRequest(http.MethodPost, listenerAddress, values)
	request.Header.Set("Origin", "http://"+listenerAddress)
	return request
}

func formRequest(method, listenerAddress string, values url.Values) *http.Request {
	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
	}
	request := httptest.NewRequest(method, "http://"+listenerAddress+"/", body)
	request.Host = listenerAddress
	if values != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return request
}

func originRequest(method, listenerAddress, origin string, values url.Values) *http.Request {
	request := formRequest(method, listenerAddress, values)
	request.Header.Set("Origin", origin)
	return request
}

func serve(t *testing.T, handler http.Handler, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", response.Header().Get("Cache-Control"))
	}
	return response
}

func assertInspectionError(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Code == http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `role="alert"`) {
		t.Fatalf("response does not contain an error: %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "Cleaned project path") {
		t.Fatalf("response contains an inspection result: %s", response.Body.String())
	}
}

func captureOutput(t *testing.T, action func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	stdout, stderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = writer, writer
	defer func() {
		os.Stdout, os.Stderr = stdout, stderr
	}()

	action()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}
