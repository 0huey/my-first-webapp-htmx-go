package main

import (
	"html/template"
	"log"
	"net/http"
	"reflect"
	"time"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

var template_glob *template.Template

type IndexContext struct {
	Count int
}

func main() {
	defer DB_Init().Close()

	template_glob = template.Must(template.ParseGlob("templates/*.html"))

	http.HandleFunc("/login", HandleLogin)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir("static")))

	http.HandleFunc("/{$}", LoginRequiredWrapper(HandleRoot))
	http.HandleFunc("/count", LoginRequiredWrapper(HandleCount))
	http.HandleFunc("/contacts", LoginRequiredWrapper(HandleContacts))

	http.Handle("/", http.NotFoundHandler())

	log.Fatal(http.ListenAndServe(":8080", LogRequests(http.DefaultServeMux)))
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case GET:
		context := IndexContext{Count: DB_GetCount()}
		Render(w, "index", context)

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func HandleCount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case POST:
		context := IndexContext{Count: DB_IncCount()}
		Render(w, "count", context)

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func Render(w http.ResponseWriter, block string, context any) {
	w.Header().Add("Cache-Control", "no-cache")
	err := template_glob.ExecuteTemplate(w, block, context)

	if err != nil {
		log.Panic(err)
	}
}

func RenderError(w http.ResponseWriter, req *http.Request, block string, context any, code int) {
	w.WriteHeader(code)
	w.Header().Add("Cache-Control", "no-cache")
	err := template_glob.ExecuteTemplate(w, block, context)

	if err != nil {
		log.Panic(err)
	}
}

func LogRequests(mux *http.ServeMux) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handler, pattern := mux.Handler(req)

		start_time := time.Now()
		handler.ServeHTTP(w, req)
		elapsed := time.Since(start_time)

		privateData := reflect.ValueOf(w).Elem()

		status := privateData.FieldByName("status").Int()
		if status == 0 {
			status = 200
		}

		log.Println(req.RemoteAddr,
			status,
			req.Method,
			req.RequestURI,
			"req-size:",
			req.ContentLength,
			"resp-size:",
			privateData.FieldByName("written"),
			"route:",
			pattern,
			elapsed,
		)
	}
}
