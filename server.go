package preseeder

import (
	"encoding/json"
	"fmt"
	// htmlTemplate "html/template"
	"log"
	"net"
	"net/http"
	"text/template"
	"time"
)

/*
 * Types
 */

type PreseedServer struct {
	preseedTemplate     *template.Template
	preseedContext      *PreseedContext
	lateCommandTemplate *template.Template
	authorizedKeys      string
	lateCommand         string
	staticFileRoot      string
	tracker             *ClientTracker
	serveMux            *http.ServeMux
}

type preseedInput struct {
	PreseedHost string
}

type serverSummary struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Urls        map[string]string `json:"urls"`
	Clients     TrackerState      `json:"clients"`
}

/*
 * Constants
 */

/*
 * Functions
 */

func (s *PreseedServer) track(name string, r *http.Request) {
	if r.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			s.tracker.Track(name, host, getRemoteMacAddress(host))
			return
		} else {
			log.Printf("track: %v", err)
		}
	}
	log.Printf("track: No remote addr found for %v", r)
}

func (s *PreseedServer) handleGetRoot(w http.ResponseWriter, r *http.Request) {
	summary := &serverSummary{
		Name:        "preseeder",
		Description: "Debian/Ubuntu preseed.cfg server. See https://github.com/hblanks/preseeder",
		Clients:     s.tracker.ReadState(),
		Urls: map[string]string{
			"preseed":      fmt.Sprintf("http://%s/preseed", r.Host),
			"late_command": fmt.Sprintf("http://%s/preseed/late_command", r.Host),
		},
	}
	summaryJSON, err := json.MarshalIndent(summary, "", "    ")
	if err == nil {
		fmt.Fprintf(w, string(summaryJSON))
	} else {
		fmt.Fprintf(w, "render failure")
	}
}

func (s *PreseedServer) handleGetPreseed(w http.ResponseWriter, r *http.Request) {
	preseedInput := s.preseedContext
	preseedInput.PreseedHost = r.Host

	err := s.preseedTemplate.Execute(w, preseedInput)
	if err != nil {
		log.Printf("handleGetPreseed: Template expansion failed: %v", err)
		http.Error(w, "Internal server error", 500)
	}
	s.track("preseed", r)
}

func (s *PreseedServer) handleGetLateCommand(w http.ResponseWriter, r *http.Request) {
	lateCommandContext := &lateCommandContext{
		s.authorizedKeys,
		s.lateCommand,
		s.preseedContext,
	}

	err := s.lateCommandTemplate.Execute(w, lateCommandContext)
	if err != nil {
		log.Printf("handleGetLateCommand: %v", err)
		http.Error(w, "Internal server error", 500)
	}
	s.track("late_command", r)
}

func (s *PreseedServer) handleFunc(
	prefix string, fn func(http.ResponseWriter, *http.Request)) {
	wrappedFn := func(w http.ResponseWriter, r *http.Request) {
		// The installer's HTTP client for fetching preseed.cfg
		// fails unless we explicitly close the HTTP connection.
		// So, we do it manually here. Alas.
		w.Header().Set("Connection", "close")
		fn(w, r)
	}
	s.serveMux.HandleFunc(prefix, wrappedFn)
}

func (s *PreseedServer) registerHandlers() {
	s.handleFunc("/preseed", s.handleGetPreseed)
	s.handleFunc("/preseed/late_command", s.handleGetLateCommand)
	if len(s.staticFileRoot) != 0 {
		staticServer := http.StripPrefix("/static/",
			http.FileServer(http.Dir(s.staticFileRoot)))
		http.Handle("/static/", staticServer)
	}
	s.handleFunc("/", s.handleGetRoot)
}

func NewPreseedServer(
	preseed string,
	preseedContext *PreseedContext,
	authorizedKeys string,
	lateCommand string,
	staticFileRoot string,
) *PreseedServer {
	if preseed == "" {
		preseed = GetDefaultPreseed()
	}

	preseedTemplate := parseTemplateString("preseed", preseed)
	lateCommandTemplate := parseTemplateString("late_command",
		GetLateCommandScript())

	s := &PreseedServer{
		preseedTemplate:     preseedTemplate,
		preseedContext:      preseedContext,
		lateCommandTemplate: lateCommandTemplate,
		authorizedKeys:      authorizedKeys,
		lateCommand:         lateCommand,
		staticFileRoot:      staticFileRoot,
		tracker:             NewClientTracker(),
		serveMux:            http.NewServeMux(),
	}
	s.registerHandlers()
	return s
}

func (s *PreseedServer) ListenAndServe(addr string, handler http.Handler) error {
	s.tracker.Start()
	httpServer := &http.Server{
		Addr:           addr,
		Handler:        s.serveMux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return httpServer.ListenAndServe()
}
