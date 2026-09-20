// Package web serves the embedded, offline study interface.
package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"

	"n400/internal/content"
)

//go:embed templates/*.gohtml static/*
var assets embed.FS

type page struct {
	Title     string
	Questions []content.Question
	Question  *content.Question
	Home      bool
	Starred   bool
	Previous  int
	Next      int
}

func New() (http.Handler, error) {
	questions, err := content.Load()
	if err != nil {
		return nil, fmt.Errorf("loading study content: %w", err)
	}
	tmpl, err := template.ParseFS(assets, "templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}
	static, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}
	render := func(w http.ResponseWriter, p page) {
		var b bytes.Buffer
		if err := tmpl.ExecuteTemplate(&b, "page", p); err != nil {
			http.Error(w, "Unable to render page", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(b.Bytes())
	}
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		render(w, page{Title: "Your civics study desk", Home: true})
	})
	mux.HandleFunc("GET /questions", func(w http.ResponseWriter, r *http.Request) {
		p := page{Title: "The 128 civics questions", Starred: r.URL.Query().Get("deck") == "6520"}
		for _, q := range questions {
			if !p.Starred || q.Is6520 {
				p.Questions = append(p.Questions, q)
			}
		}
		render(w, p)
	})
	mux.HandleFunc("GET /questions/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 || id > len(questions) {
			http.NotFound(w, r)
			return
		}
		q := questions[id-1]
		p := page{Title: fmt.Sprintf("Question %d", id), Question: &q}
		if id > 1 {
			p.Previous = id - 1
		}
		if id < len(questions) {
			p.Next = id + 1
		}
		render(w, p)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		mux.ServeHTTP(w, r)
	}), nil
}
