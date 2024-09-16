package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"myapp/data"
	"net/http"
	"time"
)

func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "login", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) PostUserLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	matches, err := user.PasswordMatches(password)
	if err != nil {
		w.Write([]byte("Error validating password"))
		return
	}

	if !matches {
		w.Write([]byte("Error validating password"))
		return
	}

	// dit the user check remember me?
	if r.Form.Get("remember") == "remember" {
		randomString := h.randomString(12)
		hasher := sha256.New()
		_, err := hasher.Write([]byte(randomString))
		if err != nil {
			h.App.ErrorInternalServerError(w, r)
			return
		}

		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		rm := data.RemeberToken{}
		err = rm.InsertToken(user.ID, sha)
		if err != nil {
			h.App.ErrorInternalServerError(w, r)
			return
		}

		// set a cookie
		cookie := http.Cookie{
			Name:     fmt.Sprintf("%s_remember", h.App.AppName),
			Value:    fmt.Sprintf("%d|%s", user.ID, sha),
			Path:     "/",
			Expires:  time.Now().Add(365 * 24 * 60 * 60 * time.Second),
			HttpOnly: true,
			Domain:   h.App.Session.Cookie.Domain,
			MaxAge:   315350000,
			Secure:   h.App.Session.Cookie.Secure,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
		// save hash in session
		h.App.Session.Put(r.Context(), "remember_token", sha)
	}

	h.sessionPut(r.Context(), "userID", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// delete the remember token if exists
	if h.App.Session.Exists(r.Context(), "remember_token") {
		rt := data.RemeberToken{}
		_ = rt.Delete(h.App.Session.GetString(r.Context(), "remember_token"))
	}

	// delete cooike
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("%s_remember", h.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   h.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   h.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)

	h.sessionRemove(r.Context(), "userID")
	h.sessionRemove(r.Context(), "remember_token")
	h.sessionDestroy(r.Context())
	h.sessionRenew(r.Context())

	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
