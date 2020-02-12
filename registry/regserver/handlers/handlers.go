// Package handlers creates HTTP handler functions for registry interface implementations
package handlers

import (
	"net/http"
	"time"

	"github.com/qri-io/qri/registry"
	"github.com/sirupsen/logrus"
)

var (
	// logger
	log = logrus.New()
)

// SetLogLevel controls how detailed handler logging is
func SetLogLevel(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}

// RouteOptions defines configuration details for NewRoutes
type RouteOptions struct {
	Protector MethodProtector
}

// AddProtector creates a configuration func for passing to NewRoutes
func AddProtector(p MethodProtector) func(o *RouteOptions) {
	return func(o *RouteOptions) {
		o.Protector = p
	}
}

// NewRoutes allocates server handlers along standard routes
func NewRoutes(reg registry.Registry, opts ...func(o *RouteOptions)) *http.ServeMux {
	o := &RouteOptions{
		Protector: NoopProtector(0),
	}
	for _, opt := range opts {
		opt(o)
	}

	pro := o.Protector
	mux := http.NewServeMux()
	mux.HandleFunc("/health", HealthCheckHandler)

	if rem := reg.Remote; rem != nil {
		// add any "/remote" routes this remote provides
		rem.AddDefaultRoutes(mux)
	}

	if ps := reg.Profiles; ps != nil {
		mux.HandleFunc("/registry/profile", logReq(NewProfileHandler(ps)))
		mux.HandleFunc("/registry/profiles", pro.ProtectMethods("POST")(logReq(NewProfilesHandler(ps))))
	}

	if s := reg.Search; s != nil {
		mux.HandleFunc("/registry/search", logReq(NewSearchHandler(s)))
	}

	return mux
}

func logReq(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	}
}

// HealthCheckHandler is a basic "hey I'm fine" for load balancers & co
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"meta":{"code": 200,"status":"ok"},"data":null}`))
}
