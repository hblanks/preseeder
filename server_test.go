package preseeder

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"text/template"
)

// TODO: consider using gocheck for assertions...

/****************** stubs for our tests' PreseedServer ****************/

func getPreseed() string {
	return `{{.PreseedHost}}`
}

func getLateCommandTemplate() *template.Template {
	return parseTemplateString("late_command", GetLateCommandScript())
}

func getAuthorizedKeys() string {
	return "ssh-rsa FOO@gar"
}

func getLateCommand() string {
	return "echo foo"
}

/******************************* Helpers ******************************/

func getTemplateOutput(t *template.Template, data interface{}) string {
	var out bytes.Buffer
	err := t.Execute(&out, data)
	if err != nil {
		log.Fatalf("TestGetPreseed: %v", err)
	}
	return out.String()
}

func serveHTTP(s *PreseedServer, w http.ResponseWriter, r *http.Request) {
	s.tracker.Start()
	s.serveMux.ServeHTTP(w, r)
	s.tracker.Stop()
}

/******************************** Tests *******************************/

func TestGetPreseed(t *testing.T) {
	preseed := getPreseed()
	preseedTemplate := parseTemplateString("preseed", preseed)

	// Get our expected output
	preseedContext := &PreseedContext{
		PreseedHost: "preseedhost:8080",
	}
	output := getTemplateOutput(preseedTemplate, preseedContext)

	// Make the request
	s := NewPreseedServer(
		preseed,
		preseedContext,
		getAuthorizedKeys(),
		getLateCommand(),
		"",
	)
	recorder := httptest.NewRecorder()
	req := &http.Request{
		Method:     "GET",
		URL:        &url.URL{Path: "/preseed"},
		Host:       preseedContext.PreseedHost,
		RemoteAddr: "10.0.14.24:23456",
	}
	serveHTTP(s, recorder, req)

	// Verify output
	if recorder.Code != 200 {
		t.Errorf("HTTP status was %v, not 200", recorder.Code)
	}
	if recorder.Body.String() != output {
		t.Errorf("HTTP status was %v, not %v",
			recorder.Body.String(), output)
	}
}

func TestGetLateCommand(t *testing.T) {
	preseed := getPreseed()
	lateCommandTemplate := getLateCommandTemplate()
	AuthorizedKeys := getAuthorizedKeys()
	LateCommand := getLateCommand()

	// Get our expected output
	input := &lateCommandInput{
		AuthorizedKeys: AuthorizedKeys,
		LateCommand:    LateCommand,
	}
	output := getTemplateOutput(lateCommandTemplate, input)

	// Make the request
	s := NewPreseedServer(
		preseed,
		&PreseedContext{},
		AuthorizedKeys,
		LateCommand,
		"",
	)
	recorder := httptest.NewRecorder()
	req := &http.Request{
		Method:     "GET",
		URL:        &url.URL{Path: "/preseed/late_command"},
		RemoteAddr: "10.0.14.24:23456",
	}
	serveHTTP(s, recorder, req)

	// Verify output
	if recorder.Code != 200 {
		t.Errorf("HTTP status was %v, not 200", recorder.Code)
	}
	if recorder.Body.String() != output {
		t.Errorf("HTTP status was %v, not %v",
			recorder.Body.String(), output)
	}
}
