// Package web serves the embedded, offline study interface.
package web

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"n400/internal/content"
	"n400/internal/flashcard"
	"n400/internal/officials"
	"n400/internal/quiz"
	"n400/internal/store"
)

//go:embed templates/*.gohtml static/*
var assets embed.FS

type page struct {
	Title           string
	Questions       []content.Question
	Question        *content.Question
	Home            bool
	Starred         bool
	Previous        int
	Next            int
	Learn           bool
	Chapters        []content.Chapter
	Chapter         *content.Chapter
	RelatedChapters []content.Chapter
	Library         bool
	LibraryDocs     []content.LibraryDoc
	LibraryDoc      *content.LibraryDoc
	RelatedLibrary  []content.LibraryDoc
	Practice        bool
	PracticeSession *practiceView
	ProgressEnabled bool
	ResolvedAnswers []officials.Answer
	Settings        bool
	States          []officials.State
	StateCode       string
	Representative  string
	ZIP             string
	District        string
	Districts       []string
	Offline         bool
	SidecarAsOf     string
	SidecarSource   string
	RefreshError    string
	GeocodeError    string
	Flashcards      bool
	Flashcard       *flashcardView
}

type practiceState struct {
	session  quiz.Session
	feedback *quiz.Result
	reviewed *content.Question
}

type practiceView struct {
	ID       string
	Session  quiz.Session
	Question *content.Question
	Feedback *quiz.Result
	Reviewed *content.Question
	Position int
}

type flashcardView struct {
	Question content.Question
	Due      int
	Total    int
	Filter   flashcard.Filter
	Empty    bool
}

func New() (http.Handler, error) {
	return newServer(nil, flashcard.SystemClock{}, "", true, officials.FederalClient{}, officials.GovernorClient{}, officials.GeocoderClient{})
}

// NewWithStore enables durable local study progress. The store is optional so
// route tests and callers that only browse content need no filesystem setup.
// sidecarPath is where a refresh writes/reads the local live-officials
// overlay ("" disables refresh and the sidecar entirely); offline disables
// the Settings refresh and address-lookup actions outright, matching the
// app's --offline flag.
func NewWithStore(progress *store.FileStore, clock flashcard.Clock, sidecarPath string, offline bool, federalClient officials.FederalClient, governorClient officials.GovernorClient, geocoderClient officials.GeocoderClient) (http.Handler, error) {
	if clock == nil {
		return nil, fmt.Errorf("flashcard clock is required")
	}
	return newServer(progress, clock, sidecarPath, offline, federalClient, governorClient, geocoderClient)
}

func newServer(progress *store.FileStore, clock flashcard.Clock, sidecarPath string, offline bool, federalClient officials.FederalClient, governorClient officials.GovernorClient, geocoderClient officials.GeocoderClient) (http.Handler, error) {
	catalog, err := content.LoadCatalog()
	if err != nil {
		return nil, fmt.Errorf("loading study content: %w", err)
	}
	questions := catalog.Questions
	snapshot, err := officials.LoadSnapshot()
	if err != nil {
		return nil, fmt.Errorf("loading officials snapshot: %w", err)
	}
	stateByCode := func(code string) officials.State {
		for _, state := range snapshot.States {
			if state.Code == code {
				return state
			}
		}
		return officials.State{}
	}
	resolver := officials.Resolver{Snapshot: snapshot}
	resolverFor := func(profile store.Profile) officials.Resolver {
		active := resolver
		active.Manual = profile.Overrides
		active.District = profile.District
		if sidecarPath != "" {
			if sidecar, err := officials.LoadSidecar(sidecarPath); err == nil {
				active.Sidecar = sidecar.Overrides(profile.State)
				active.SidecarAsOf = sidecar.AsOf
			}
		}
		return active
	}
	scheduler := flashcard.NewScheduler(clock)
	var practiceMu sync.Mutex
	practiceSessions := map[string]*practiceState{}
	nextPracticeID := 0
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
	mux.HandleFunc("GET /images/{file}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		b, err := catalog.ImageBytes(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(b))
	})
	mux.HandleFunc("GET /learn", func(w http.ResponseWriter, r *http.Request) {
		render(w, page{Title: "Learn the story behind the answers", Learn: true, Chapters: catalog.Chapters})
	})
	mux.HandleFunc("GET /learn/{chapter}", func(w http.ResponseWriter, r *http.Request) {
		for _, chapter := range catalog.Chapters {
			if chapter.ID != r.PathValue("chapter") {
				continue
			}
			p := page{Title: chapter.Title, Chapter: &chapter}
			for _, id := range chapter.Questions {
				p.Questions = append(p.Questions, questions[id-1])
			}
			render(w, p)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /library", func(w http.ResponseWriter, r *http.Request) {
		render(w, page{Title: "The founding documents and more", Library: true, LibraryDocs: catalog.Library})
	})
	mux.HandleFunc("GET /library/{doc}", func(w http.ResponseWriter, r *http.Request) {
		for _, doc := range catalog.Library {
			if doc.ID != r.PathValue("doc") {
				continue
			}
			p := page{Title: doc.Title, LibraryDoc: &doc}
			for _, id := range doc.Questions {
				p.Questions = append(p.Questions, questions[id-1])
			}
			render(w, p)
			return
		}
		http.NotFound(w, r)
	})
	flashcardPage := func(w http.ResponseWriter, r *http.Request) {
		filter := flashcard.Filter{Section: r.URL.Query().Get("section"), Chapter: r.URL.Query().Get("chapter"), Only6520: r.URL.Query().Get("deck") == "6520", Missed: r.URL.Query().Get("missed") == "1"}
		data := store.Empty()
		if progress != nil {
			loaded, err := progress.Load()
			if err == nil {
				data = loaded
			}
		}
		missed := make(map[int]bool, len(data.Questions))
		for id, record := range data.Questions {
			missed[id] = record.Incorrect > record.Correct
		}
		deck := flashcard.Build(questions, filter, missed)
		view := &flashcardView{Total: len(deck), Filter: filter}
		for _, question := range deck {
			card, seen := data.Cards[question.ID]
			if scheduler.Due(card, seen) {
				view.Due++
				if view.Question.ID == 0 {
					view.Question = question
				}
			}
		}
		view.Empty = view.Question.ID == 0
		render(w, page{Title: "Flashcards", Flashcards: true, Flashcard: view})
	}
	mux.HandleFunc("GET /flashcards", flashcardPage)
	mux.HandleFunc("POST /flashcards/{id}", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Flashcard progress is unavailable", http.StatusServiceUnavailable)
			return
		}
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 || id > len(questions) {
			http.NotFound(w, r)
			return
		}
		correct := r.FormValue("result") == "right"
		if r.FormValue("result") != "right" && r.FormValue("result") != "wrong" {
			http.Error(w, "Choose whether you recalled the answer", http.StatusBadRequest)
			return
		}
		if err := progress.Update(func(data *store.Data) {
			card := data.Cards[id]
			data.Cards[id] = scheduler.Review(card, correct)
			record := data.Questions[id]
			if correct {
				record.Correct++
			} else {
				record.Incorrect++
			}
			record.LastAnswered = clock.Now()
			data.Questions[id] = record
		}); err != nil {
			http.Error(w, "Unable to save flashcard progress", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/flashcards", http.StatusSeeOther)
	})
	settingsPage := func(refreshError, geocodeError string) page {
		stored, err := progress.Load()
		if err != nil {
			stored = store.Empty()
		}
		p := page{Title: "Settings", Settings: true, States: snapshot.States, StateCode: stored.Profile.State, Representative: stored.Profile.Overrides[29], ZIP: stored.Profile.ZIP, District: stored.Profile.District, Districts: snapshot.DistrictCandidates(stored.Profile.ZIP), Offline: offline, RefreshError: refreshError, GeocodeError: geocodeError}
		if sidecarPath != "" {
			if sidecar, err := officials.LoadSidecar(sidecarPath); err == nil {
				p.SidecarAsOf, p.SidecarSource = sidecar.AsOf, sidecar.Source
			}
		}
		return p
	}
	mux.HandleFunc("GET /settings", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		render(w, settingsPage("", ""))
	})
	mux.HandleFunc("POST /settings/state", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		code := strings.ToUpper(strings.TrimSpace(r.FormValue("state")))
		valid := false
		for _, state := range snapshot.States {
			if state.Code == code {
				valid = true
				break
			}
		}
		if !valid {
			http.Error(w, "Choose a state", http.StatusBadRequest)
			return
		}
		if err := progress.Update(func(data *store.Data) { data.Profile.State = code }); err != nil {
			http.Error(w, "Unable to save settings", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	})
	mux.HandleFunc("POST /settings/representative", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		name := strings.TrimSpace(r.FormValue("representative"))
		if name == "" {
			http.Error(w, "Enter your representative's name", http.StatusBadRequest)
			return
		}
		if err := progress.Update(func(data *store.Data) { data.Profile.Overrides[29] = name }); err != nil {
			http.Error(w, "Unable to save settings", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	})
	mux.HandleFunc("POST /settings/zip", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		zip := strings.TrimSpace(r.FormValue("zip"))
		if len(zip) != 5 {
			http.Error(w, "Enter a five-digit ZIP code", http.StatusBadRequest)
			return
		}
		for _, digit := range zip {
			if digit < '0' || digit > '9' {
				http.Error(w, "Enter a five-digit ZIP code", http.StatusBadRequest)
				return
			}
		}
		candidates := snapshot.DistrictCandidates(zip)
		if len(candidates) == 0 {
			http.Error(w, "ZIP code was not found in the bundled district crosswalk", http.StatusNotFound)
			return
		}
		if err := progress.Update(func(data *store.Data) {
			data.Profile.ZIP = zip
			data.Profile.District = ""
			if len(candidates) == 1 {
				data.Profile.District = candidates[0]
			}
		}); err != nil {
			http.Error(w, "Unable to save settings", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	})
	mux.HandleFunc("POST /settings/district", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		district := strings.TrimSpace(r.FormValue("district"))
		stored, err := progress.Load()
		if err != nil {
			http.Error(w, "Unable to load settings", http.StatusInternalServerError)
			return
		}
		valid := false
		for _, candidate := range snapshot.DistrictCandidates(stored.Profile.ZIP) {
			if candidate == district {
				valid = true
				break
			}
		}
		if !valid {
			http.Error(w, "Choose one of the ZIP's district candidates", http.StatusBadRequest)
			return
		}
		if err := progress.Update(func(data *store.Data) { data.Profile.District = district }); err != nil {
			http.Error(w, "Unable to save settings", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	})
	mux.HandleFunc("POST /settings/refresh", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		if offline || sidecarPath == "" {
			http.Error(w, "Refresh is disabled while the app is running offline", http.StatusServiceUnavailable)
			return
		}
		stored, err := progress.Load()
		if err != nil {
			stored = store.Empty()
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if _, err := officials.Refresh(ctx, federalClient, governorClient, stateByCode(stored.Profile.State), sidecarPath, clock.Now()); err != nil {
			render(w, settingsPage(err.Error(), ""))
			return
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	})
	mux.HandleFunc("POST /settings/address", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		if offline {
			http.Error(w, "Address lookup is disabled while the app is running offline", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		district, err := geocoderClient.Fetch(ctx, r.FormValue("address"))
		if err != nil {
			render(w, settingsPage("", err.Error()))
			return
		}
		if err := progress.Update(func(data *store.Data) { data.Profile.District = district }); err != nil {
			http.Error(w, "Unable to save settings", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	})
	mux.HandleFunc("POST /settings/reset-officials", func(w http.ResponseWriter, r *http.Request) {
		if progress == nil {
			http.Error(w, "Settings are unavailable", http.StatusServiceUnavailable)
			return
		}
		if sidecarPath != "" {
			if err := officials.ResetSidecar(sidecarPath); err != nil {
				http.Error(w, "Unable to reset officials data", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	})
	practicePage := func(w http.ResponseWriter, id string, state *practiceState) {
		view := &practiceView{ID: id, Session: state.session, Feedback: state.feedback, Reviewed: state.reviewed}
		if !state.session.Complete {
			question := state.session.Questions[state.session.Answered]
			view.Question = &question
			view.Position = state.session.Answered + 1
		} else {
			view.Position = state.session.Answered
		}
		render(w, page{Title: "Practice test", Practice: true, PracticeSession: view, ProgressEnabled: progress != nil})
	}
	mux.HandleFunc("GET /practice", func(w http.ResponseWriter, r *http.Request) {
		render(w, page{Title: "Practice test", Practice: true, ProgressEnabled: progress != nil})
	})
	mux.HandleFunc("POST /practice/start", func(w http.ResponseWriter, r *http.Request) {
		mode := quiz.Mode(r.FormValue("mode"))
		if mode == "" {
			mode = quiz.Official
		}
		session, err := quiz.NewSession(questions, mode, rand.New(rand.NewSource(time.Now().UnixNano())))
		if err != nil {
			http.Error(w, "Unable to start practice test", http.StatusBadRequest)
			return
		}
		practiceMu.Lock()
		nextPracticeID++
		id := strconv.Itoa(nextPracticeID)
		practiceSessions[id] = &practiceState{session: session}
		practiceMu.Unlock()
		http.Redirect(w, r, "/practice/"+id, http.StatusSeeOther)
	})
	mux.HandleFunc("GET /practice/{id}", func(w http.ResponseWriter, r *http.Request) {
		practiceMu.Lock()
		state := practiceSessions[r.PathValue("id")]
		if state != nil {
			practicePage(w, r.PathValue("id"), state)
		}
		practiceMu.Unlock()
		if state == nil {
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("POST /practice/{id}", func(w http.ResponseWriter, r *http.Request) {
		practiceMu.Lock()
		defer practiceMu.Unlock()
		state := practiceSessions[r.PathValue("id")]
		if state == nil {
			http.NotFound(w, r)
			return
		}
		if state.session.Complete {
			practicePage(w, r.PathValue("id"), state)
			return
		}
		question := state.session.Questions[state.session.Answered]
		stateCode := ""
		active := resolver
		if progress != nil {
			if data, err := progress.Load(); err == nil {
				stateCode = data.Profile.State
				active = resolverFor(data.Profile)
			}
		}
		resolved := active.ResolveAll(question.ID, stateCode)
		values := make([]string, 0, len(resolved))
		for _, answer := range resolved {
			values = append(values, answer.Text)
		}
		result := quiz.GradeWithResolvedAnswers(question, r.FormValue("answer"), values)
		correct := result.Correct || r.FormValue("override") == "right"
		if r.FormValue("override") == "right" {
			result.Correct = true
			result.Message = "Marked right by you."
		}
		if err := state.session.Record(correct); err != nil {
			http.Error(w, "Unable to record answer", http.StatusConflict)
			return
		}
		if progress != nil {
			if err := progress.Update(func(data *store.Data) {
				record := data.Questions[question.ID]
				if correct {
					record.Correct++
				} else {
					record.Incorrect++
				}
				record.LastAnswered = clock.Now()
				data.Questions[question.ID] = record
			}); err != nil {
				http.Error(w, "Unable to save practice progress", http.StatusInternalServerError)
				return
			}
		}
		state.feedback = &result
		state.reviewed = &question
		practicePage(w, r.PathValue("id"), state)
	})
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
		if q.Changing() {
			stateCode := ""
			active := resolver
			if progress != nil {
				if data, err := progress.Load(); err == nil {
					stateCode = data.Profile.State
					active = resolverFor(data.Profile)
				}
			}
			p.ResolvedAnswers = active.ResolveAll(q.ID, stateCode)
		}
		for _, chapterID := range q.Chapters {
			for _, chapter := range catalog.Chapters {
				if chapter.ID == chapterID {
					p.RelatedChapters = append(p.RelatedChapters, chapter)
				}
			}
		}
		for _, docID := range q.Library {
			for _, doc := range catalog.Library {
				if doc.ID == docID {
					p.RelatedLibrary = append(p.RelatedLibrary, doc)
				}
			}
		}
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
