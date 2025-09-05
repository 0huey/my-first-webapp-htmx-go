package main

import(
	"net/http"
	"html/template"
	"log"
	"reflect"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type IndexContext struct {
	Count int
}

func main() {
	defer DB_Init().Close()

	http.HandleFunc("/{$}", HandleRoot)
	http.HandleFunc("/count", HandleCount)

	http.HandleFunc("/contacts", HandleContacts)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir("static")))

	http.Handle("/", http.NotFoundHandler())

	log.Fatal(http.ListenAndServe(":8080", LogRequests(http.DefaultServeMux)))
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case GET:
			context := IndexContext{ Count: DB_GetCount() }
			Render(w, "index", context)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func HandleCount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case POST:
			context := IndexContext{ Count: DB_IncCount() }
			Render(w, "count", context)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func Render(w http.ResponseWriter, block string, context any) {
	t := template.Must(template.ParseGlob("templates/*.html"))

	w.Header().Add("Cache-Control", "no-cache")

	err := t.ExecuteTemplate(w, block, context)

	if err != nil {
		log.Println(err)
	}
}

func RenderError(w http.ResponseWriter, req *http.Request, block string, context any, code int) {
	w.WriteHeader(code)
	w.Header().Add("Cache-Control", "no-cache")

	t := template.Must(template.ParseGlob("templates/*.html"))

	err := t.ExecuteTemplate(w, block, context)

	if err != nil {
		log.Println(err)
	}
}

func LogRequests(mux *http.ServeMux) http.HandlerFunc {
	return func (w http.ResponseWriter, req *http.Request) {

		handler, pattern := mux.Handler(req)
		handler.ServeHTTP(w, req)

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
		)
	}
}
