package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

//go:embed templates/index.html
var templatesFS embed.FS

const description = "This site is under construction. The gophers are hard at work building something great. Please check back soon!"

// sessionCookieName identifies a returning visitor so reloads aren't counted
// as new visits.
const sessionCookieName = "cz_session"

// newSessionToken returns a random hex token for a visitor's session cookie.
func newSessionToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// gopherASCII is a small ASCII rendering of the Go gopher mascot.
// NOTE: keep this free of backtick characters so the raw string literal stays valid.
const gopherASCII = `         ,_---~~~~~----._
   _,,_,*^____      _____*g*"*,
  / __/ /'     ^.  /      \ ^@q   f
 [  @f | @))    |  | @))   l  0 _/
  \ /   \~____ / __ \_____/    \
   |           _l__l_           I
   }          [______]          I
   ]            | | |           |
   ]             ~ ~            |
   |                           |
    \                         /
     ^,_                   _,^
        ^~--___________--~^`

type pageData struct {
	Description  string
	Gopher       string
	VisitorCount uint64
	Address      string
	Uptime       string
	CurrentTime  string
	GoVersion    string
	GitHubURL    string
	// StartUnix and ServerNowUnix let the client tick the uptime/clock every
	// second while staying aligned to the server's clock.
	StartUnix     int64
	ServerNowUnix int64
}

const defaultGitHubURL = "https://github.com/elug3/construct-zone"

func githubURL() string {
	if u := os.Getenv("GITHUB_URL"); u != "" {
		return u
	}
	return defaultGitHubURL
}

var (
	startTime    = time.Now()
	visitorCount uint64
	tmpl         = template.Must(template.ParseFS(templatesFS, "templates/index.html"))
)

func humanizeUptime(d time.Duration) string {
	d = d.Round(time.Second)
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second

	var b strings.Builder
	if days > 0 {
		writeUnit(&b, int(days), "day")
	}
	if hours > 0 {
		writeUnit(&b, int(hours), "hour")
	}
	if minutes > 0 {
		writeUnit(&b, int(minutes), "minute")
	}
	writeUnit(&b, int(seconds), "second")
	return strings.TrimSpace(b.String())
}

func writeUnit(b *strings.Builder, value int, unit string) {
	if b.Len() > 0 {
		b.WriteString(" ")
	}
	b.WriteString(itoa(value))
	b.WriteString(" ")
	b.WriteString(unit)
	if value != 1 {
		b.WriteString("s")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Only serve the page for the root path so we don't double-count assets.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Count a visit only when there's no existing session cookie, so reloads by
	// the same visitor don't inflate the count. New visitors get a session
	// cookie and bump the unique-visitor total.
	var count uint64
	if _, err := r.Cookie(sessionCookieName); err != nil {
		token, terr := newSessionToken()
		if terr != nil {
			log.Printf("failed to generate session token: %v", terr)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   60 * 60 * 24 * 365,
		})
		count = atomic.AddUint64(&visitorCount, 1)
	} else {
		count = atomic.LoadUint64(&visitorCount)
	}

	now := time.Now()
	data := pageData{
		Description:   description,
		Gopher:        gopherASCII,
		VisitorCount:  count,
		Address:       r.Host,
		Uptime:        humanizeUptime(now.Sub(startTime)),
		CurrentTime:   now.UTC().Format("Mon, 02 Jan 2006 15:04:05 MST"),
		GoVersion:     "Go " + strings.TrimPrefix(runtime.Version(), "go"),
		GitHubURL:     githubURL(),
		StartUnix:     startTime.Unix(),
		ServerNowUnix: now.Unix(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// The page is dynamic (live counters, per-request stats), so never cache it.
	w.Header().Set("Cache-Control", "no-store")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/healthz", healthHandler)

	addr := ":" + port
	log.Printf("Construct-Zone listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
