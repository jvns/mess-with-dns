package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func createRoutes(handle *handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /records", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		getRecords(username, handle.rs, w, r)
	}))
	mux.Handle("DELETE /records/{record_id}", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		recordID := r.PathValue("record_id")
		username := r.Context().Value("username").(string)
		deleteRecord(username, recordID, handle.rs, w, r)
	}))
	mux.Handle("DELETE /records", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		deleteAllRecords(username, handle.rs, w, r)
	}))
	mux.Handle("POST /records/{record_id}", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		recordID := r.PathValue("record_id")
		username := r.Context().Value("username").(string)
		updateRecord(username, recordID, handle.rs, w, r)
	}))
	mux.Handle("POST /records", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		createRecord(username, handle.rs, w, r)
	}))
	mux.Handle("GET /requests", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		getRequests(handle.logger, username, w, r)
	}))
	mux.Handle("DELETE /requests", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		deleteRequests(handle.logger, username, w, r)
	}))
	mux.Handle("GET /requeststream/{username}", handle.addMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		if username == "" {
			username = r.PathValue("username")
		}
		streamRequests(handle.logger, username, w, r)
	}))
	mux.Handle("GET /login/", addBaseMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		loginRandom(handle.userService, handle.rs, w, r)
	}))
	mux.Handle("GET /health", addBaseMiddlewares(healthCheck))
	mux.Handle("GET /health/", addBaseMiddlewares(healthCheck))
	mux.Handle("GET /", addBaseMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=120")
		http.ServeFile(w, r, handle.workdir+"/frontend/"+r.URL.Path)
	}))

	return mux
}

func addBaseMiddlewares(handleFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return otelhttp.NewHandler(addDotNetMiddleware(http.HandlerFunc(handleFunc)), "mess-with-dns-api")
}

func (handle *handler) addMiddlewares(handleFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return otelhttp.NewHandler(
		addDotNetMiddleware(
			handle.corsLoginMiddleware(
				http.HandlerFunc(handleFunc))), "mess-with-dns-api")
}

func (handle *handler) corsLoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		username, _ := handle.userService.ReadSessionUsername(r)

		// tracing
		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(attribute.String("http.path", r.URL.Path))
		span.SetAttributes(attribute.String("username", username))

		// disable caching
		w.Header().Set("Cache-Control", "no-store")

		// check that user is logged in
		ctx := context.WithValue(r.Context(), "username", username)
		r = r.WithContext(ctx)
		logMsg(r, fmt.Sprintf("%s %s (%s)", r.Method, r.URL.Path, username))
		if username == "" {
			page := r.URL.Path
			returnError(w, r, fmt.Errorf("you must be logged in to access this page: %s", page), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func addDotNetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "messwithdns.com" || r.Host == "www.messwithdns.com" {
			http.Redirect(w, r, "https://messwithdns.net"+r.URL.Path, http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
