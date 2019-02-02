package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func setRouting(token string, db IStorage) *mux.Router {
	r := mux.NewRouter()
	r.Path("/api-01/task/{task}").Methods("GET").HandlerFunc(
		makeHandlerWithStore(taskInfoHandler, db))
	r.Path("/api-01/task/{task}").Methods("OPTIONS").HandlerFunc(optionsHandler)
	r.Path("/api-01/task/{task}/ok").Methods("PATCH").HandlerFunc(
		superTokenAuth(makeHandlerWithStoreAndParam(taskCompleteHandler, db, "ok"), token))
	r.Path("/api-01/task/{task}/fail").Methods("PATCH").HandlerFunc(
		superTokenAuth(makeHandlerWithStoreAndParam(taskCompleteHandler, db, "fail"), token))
	r.Path("/api-01/upload").Methods("POST").HandlerFunc(
		makeHandlerWithStore(uploadHandler, db))
	r.Path("/api-01/upload").Methods("OPTIONS").HandlerFunc(optionsHandler)
	r.Path("/api-01/queue").Methods("GET").HandlerFunc(
		superTokenAuth(makeHandlerWithStore(queueFirstHandler, db), token))
	r.PathPrefix("/").HandlerFunc(invalidRequest)
	return r
}

func makeHandlerWithStore(fn func(http.ResponseWriter, *http.Request, IStorage), db IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db)
	}
}

func makeHandlerWithStoreAndParam(fn func(http.ResponseWriter, *http.Request, IStorage, string), db IStorage, p string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db, p)
	}
}

func invalidRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sendJSONErrorMessage(w, E_INVALID_REQUEST, http.StatusBadRequest)
	Warning.Printf("(-) [%s]: Unknown path: %s\n", r.RemoteAddr, r.RequestURI)
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sendJSONErrorMessage(w, E_NOT_IMPLEMENTED, http.StatusSeeOther)
	Warning.Printf("(-) [%s]: Not implemented yet: %s", r.RemoteAddr, r.RequestURI)
}

func superTokenAuth(fn http.HandlerFunc, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// BearerAuth returns the token provided in the request's
		// Authorization header, if the request uses HTTP Bearer Authentication.
		// BearerAuth parses an HTTP Basic Authentication string.
		// "Bearer QWxhZGRpbjpvcGVuIHNlc2FtZQ" returns ("QWxhZGRpbjpvcGVuIHNlc2FtZQ", true).
		var _token string
		auth := r.Header.Get("Authorization")
		if auth != "" {
			const prefix = "Bearer "
			if strings.HasPrefix(auth, prefix) {
				if len(prefix) < len(auth) {
					_token = string(auth[len(prefix):])
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if _token == "" || token != _token {
			w.Header().Add("WWW-Authenticate", "Basic")
			sendJSONErrorMessage(w, E_ACCESS_DENIED, http.StatusUnauthorized)
			Warning.Printf("(%s) [%s]: Super Client authentication failed", token, r.RemoteAddr)
			return
		}
		fn(w, r)
	}
}
