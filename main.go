package main

import(
	"net/http"
	"html/template"
	"log"
	"strconv"
	"reflect")

type Contact struct {
	Id int
	Name string
	Email string
}
type ContactSlice []Contact

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type ClientContext struct {
	Count int
	Contacts ContactSlice
	FormErrors FormData
}

var contact_id int = 0

var context ClientContext

func main() {
	// some initial data
	context.Contacts.AddContact(NewContact("john smith", "js@example.com"))
	context.Contacts.AddContact(NewContact("person", "person@example.com"))

	httpMux := http.NewServeMux()

	httpMux.HandleFunc("/{$}", HandleRoot)
	httpMux.HandleFunc("/count", HandleCount)
	httpMux.HandleFunc("/contacts", HandleContact)

	httpMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	httpMux.Handle("/favicon.ico", http.FileServer(http.Dir("static")))

	httpMux.Handle("/", http.NotFoundHandler())

	log.Fatalln(http.ListenAndServe(":8080", LogRequests(httpMux)))
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "GET":
		Render(w, req, "index", context)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func HandleCount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "POST":
		context.Count++
		Render(w, req, "count", context)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func HandleContact(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "POST": {
			name  := req.PostFormValue("name")
			email := req.PostFormValue("email")

			if len(name) == 0 || len(email) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if context.Contacts.EmailExists(email) {
				temp_context := context
				temp_context.FormErrors = newFormData()
				temp_context.FormErrors.Values["name"] = name
				temp_context.FormErrors.Values["email"] = email
				temp_context.FormErrors.Errors["message"] = "that email address already exists retard"
				RenderError(w, req, "email form", temp_context, http.StatusConflict)
				return
			}

			new := NewContact(name, email)
			context.Contacts.AddContact(new)
			Render(w, req, "oob-contact", new)
			Render(w, req, "email form", context)
		}

		case "DELETE": {
			id, err := strconv.Atoi(req.URL.Query().Get("id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			for i, c := range context.Contacts {
				if c.Id == id {
					context.Contacts = context.Contacts[:i+copy(context.Contacts[i:], context.Contacts[i+1:])]
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func Render(w http.ResponseWriter, req *http.Request, block string, context any) {
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

func LogRequests(httpMux *http.ServeMux) http.HandlerFunc {
	return func (w http.ResponseWriter, req *http.Request) {

		handler, pattern := httpMux.Handler(req)
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

func NewContact(name string, email string) Contact {
	contact_id++
	return Contact {Id: contact_id,
		Name: name,
		Email: email,
	}
}

func (c *ContactSlice) AddContact(new Contact) {
	*c = append(*c, new)
}

func (c ContactSlice) EmailExists(email string) bool {
	for _, contact := range c {
		if contact.Email == email {
			return true
		}
	}
	return false
}
