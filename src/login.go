package main

import (
	"fmt"
	_ "log"
	"net/http"
	"time"
)

type LoginSession struct {
	Username string
	Token    string
	Expires  time.Time
}

type LoginPageData struct {
	Username string
	Error    string
}

func HandleLogin(w http.ResponseWriter, req *http.Request) {
	var data LoginPageData

	switch req.Method {
	case GET:
		{
			session, err := req.Cookie("session")
			if err == nil {
				username, err := DB_UserLookupLoginToken(session.Value)

				if err == nil {
					data.Username = username

				} else if err == DB_ErrorLoginSessionExpired {
					data.Error = "Your session has expired. Please log in again"
					w.Header().Add("Set-Cookie", "session=; expires=Thu, 01 Jan 1970 00:00:00 GMT")
				}
			}
			Render(w, "login-page", data)
		}

	case POST:
		{
			username := req.PostFormValue("username")
			password := req.PostFormValue("password")

			token, err := DB_UserLogin(username, password)

			if err == nil {
				data.Username = token.Username
				cookie := fmt.Sprintf("session=%s; Max-Age=3600", token.Token)
				w.Header().Add("Set-Cookie", cookie)
				fmt.Println(data)
				Render(w, "login-page-username", data)

			} else if err == DB_ErrorAccountLocked {
				data.Error = "Your account had been disabled. Please contact your administrator"
				Render(w, "login-page-form", data)

			} else {
				data.Error = "Unknown username or password"
				Render(w, "login-page-form", data)
			}
		}

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func LoginRequiredWrapper(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		session, err := req.Cookie("session")

		if err != nil {
			http.Redirect(w, req, "/login", http.StatusFound)
			return
		}

		_, err = DB_UserLookupLoginToken(session.Value)

		if err != nil {
			http.Redirect(w, req, "/login", http.StatusFound)
		}

		handler.ServeHTTP(w, req)
	}
}
