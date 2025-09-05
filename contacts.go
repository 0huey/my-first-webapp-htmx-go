package main

import (
	"net/http"
	"log"
	"strconv"
	"strings"
	"regexp"
)

type ContactEntry struct {
	Id int
	Name string
	Email string
}

type ContactList struct {
	Contacts []ContactEntry
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func HandleContacts(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case GET: {
			str_id := req.URL.Query().Get("id")
			str_edit := req.URL.Query().Get("edit")

			if len(str_id) > 0 {
				//GET 1 contact

				id, err := strconv.Atoi(str_id)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				contact := DB_GetOneContact(id)
				Render(w, "contact-row", contact)

			} else if len(str_edit) > 0 {
				// send edit form

				edit_id, err := strconv.Atoi(str_edit)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				contact := DB_GetOneContact(edit_id)
				form := contact.ToFormData()
				Render(w, "contact-edit-form", form)

			} else {
				all := ContactList{Contacts: DB_GetAllContacts()}
				Render(w, "contacts-page", all)
			}
		}

		case POST: {
			new := ContactEntry {
				Name:  req.PostFormValue("name"),
				Email: req.PostFormValue("email"),
			}

			if len(new.Name) == 0 || len(new.Email) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			validateAndInsertNewContact(new, w, req)
		}

		case PUT: {
			id, err := strconv.Atoi(req.URL.Query().Get("id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			new := ContactEntry {Id: id,
				Name:  req.PostFormValue("name"),
				Email: req.PostFormValue("email"),
			}
			validateAndInsertNewContact(new, w, req)
		}

		case DELETE: {
			id, err := strconv.Atoi(req.URL.Query().Get("id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			DB_DeleteContact(id)
			w.WriteHeader(http.StatusOK)
		}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func validateAndInsertNewContact(new ContactEntry, w http.ResponseWriter, req *http.Request) {
	if len(new.Name) == 0 || len(new.Email) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	new.Name  = strings.TrimSpace(new.Name)
	new.Email = strings.TrimSpace(new.Email)
	new.Email = strings.ToLower(new.Email)

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
			DB_AddContact(new)
			Render(w, "oob-contact-add", new)
			Render(w, "new-contact-form-blank", newFormData())
		case PUT:
			DB_UpdateContact(new)
			Render(w, "contact-row", new)
		default:
			log.Println("ERROR unknown method", req.Method, "in validateAndInsertNewContact")
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
