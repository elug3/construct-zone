package main

import (
	"embed"
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
	Domain       string
	Uptime       string
	CurrentTime  string
	GoVersion    string
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

	count := atomic.AddUint64(&visitorCount, 1)

	data := pageData{
		Description:  description,
		Gopher:       gopherASCII,
		VisitorCount: count,
		Domain:       r.Host,
		Uptime:       humanizeUptime(time.Since(startTime)),
		CurrentTime:  time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 MST"),
		GoVersion:    "Go " + strings.TrimPrefix(runtime.Version(), "go"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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
