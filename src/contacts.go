package main

import (
	"html"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type ContactEntry struct {
	Id    int
	Name  string
	Email string
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func HandleContacts(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case GET:
		{
			params := req.URL.Query()

			if params.Has("id") {
				//GET 1 contact

				id, err := strconv.Atoi(params.Get("id"))
				if err != nil {
					http.Error(w, "Malformed ID", http.StatusBadRequest)
					return
				}
				contact := DB_GetOneContact(id)
				if contact.IsNull() {
					http.NotFound(w, req)
					return
				}
				Render(w, "contact-row", contact)

			} else if params.Has("edit") {
				// send edit form

				edit_id, err := strconv.Atoi(params.Get("edit"))
				if err != nil {
					http.Error(w, "Malformed ID", http.StatusBadRequest)
					return
				}
				contact := DB_GetOneContact(edit_id)
				if contact.IsNull() {
					http.NotFound(w, req)
					return
				}
				form := contact.ToFormData()
				Render(w, "contact-edit-form", form)

			} else if params.Has("search") {
				Render(w, "contacts-table", DB_GetAllContactsSearch(params.Get("search")))

			} else {
				Render(w, "contacts-page", nil)
			}
		}

	case POST:
		{
			new := ContactEntry{
				Name:  req.PostFormValue("name"),
				Email: req.PostFormValue("email"),
			}

			if len(new.Name) == 0 || len(new.Email) == 0 {
				http.Error(w, "Malformed data", http.StatusBadRequest)
				return
			}
			validateAndInsertNewContact(new, w, req)
		}

	case PUT:
		{
			id, err := strconv.Atoi(req.URL.Query().Get("id"))
			if err != nil {
				http.Error(w, "Malformed ID", http.StatusBadRequest)
				return
			}
			new := ContactEntry{Id: id,
				Name:  req.PostFormValue("name"),
				Email: req.PostFormValue("email")}
			validateAndInsertNewContact(new, w, req)
		}

	case DELETE:
		{
			id, err := strconv.Atoi(req.URL.Query().Get("id"))
			if err != nil {
				http.Error(w, "Malformed ID", http.StatusBadRequest)
				return
			}

			DB_DeleteContact(id)
			// status header will default to OK
			return
		}

	default:
		http.Error(w, "Malformed ID", http.StatusBadRequest)
		return
	}
}

func validateAndInsertNewContact(new ContactEntry, w http.ResponseWriter, req *http.Request) {
	if len(new.Name) == 0 || len(new.Email) == 0 {
		http.Error(w, "Malformed data", http.StatusBadRequest)
		return
	}

	new.Name = strings.TrimSpace(new.Name)
	new.Name = html.EscapeString(new.Name)

	new.Email = strings.TrimSpace(new.Email)
	new.Email = strings.ToLower(new.Email)
	new.Email = html.EscapeString(new.Email)

	form := new.ToFormData()

	re_email := regexp.MustCompile("[^@]+@[^@]")

	if !re_email.MatchString(new.Email) {
		form.Errors["Message"] = "invalid email address format"

		switch req.Method {
		case POST:
			RenderError(w, req, "new-contact-form-with-error", form, http.StatusConflict)
		case PUT:
			RenderError(w, req, "contact-edit-form", form, http.StatusConflict)
		default:
			log.Println("ERROR unknown method", req.Method, "in validateAndInsertNewContact")
		}
		return
	}

	if req.Method == POST && DB_ContactEmailExists(new) {
		form.Errors["Message"] = "that email address already exists"
		RenderError(w, req, "new-contact-form-with-error", form, http.StatusConflict)
		return
	}

	switch req.Method {
	case POST:
		new = DB_AddContact(new)
		Render(w, "oob-contact-add", new)
		Render(w, "new-contact-form-blank", newFormData())
	case PUT:
		DB_UpdateContact(new)
		Render(w, "contact-row", new)
	default:
		log.Println("ERROR unknown method", req.Method, "in validateAndInsertNewContact")
		return
	}
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

func (c ContactEntry) ToFormData() FormData {
	d := newFormData()
	d.Values["Id"] = strconv.Itoa(c.Id)
	d.Values["Name"] = c.Name
	d.Values["Email"] = c.Email
	return d
}

func (c ContactEntry) IsNull() bool {
	return c.Id == 0
}
