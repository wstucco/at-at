package at_at

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultConfigFile = ".at_at"
	DefaultPort       = 80
	DefaultHost       = "0.0.0.0"
)

// module reference to the router
var router *Router

type Router struct {
	hosts HostList
}

func init() {
	router = &Router{}
}

func NewRouter(hosts HostList) *Router {
	router.hosts = hosts
	return router
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if host := r.hosts[stripHostTLD(req.Host)]; host != nil {

		if host.Error != NoError {
			handleHostError(res, host)
		} else if host.status == Stopped {
			host.Run()
		}

		if host.status == Running {
			host.ServeHTTP(res, req)
		}

		return
	}

	handleNotFound(res, &Host{name: stripHostTLD(req.Host)})
}

func (r *Router) Run() {
	address := fmt.Sprintf("%s:%d", host(), port())
	Logger().Printf("At-At is listening on %s\n", address)
	err := http.ListenAndServe(address, router)
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(address, r)
}

func port() int {
	port := getVar("port")

	if port == nil {
		return DefaultPort
	}

	if i, ok := port.(int); ok {
		return i
	}

	if s, ok := port.(string); ok {
		i, err := strconv.Atoi(s)
		if err != nil {
			return DefaultPort
		}

		return i
	}

	return DefaultPort
}

func host() string {
	host := getStringVarWithDefault("host", DefaultHost)

	return host
}

func stripHostTLD(host string) string {
	// this just remove the extension, like it was a file name
	// NEED FIX
	return host[0 : len(host)-len(filepath.Ext(host))]
}

func getVar(key string) interface{} {
	envKey := fmt.Sprintf("AT_AT_%s", strings.ToUpper(key))
	if value := os.Getenv(envKey); value != "" {
		return value
	}

	return nil
}

func getStringVar(key string) string {
	value := getVar(key)
	if s, ok := value.(string); ok {
		return s
	}

	return ""
}

func getStringVarWithDefault(key, defaultValue string) string {
	value := getStringVar(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func handleHostError(w http.ResponseWriter, host *Host) {
	switch host.Error {
	case Unavailable:
		handleNotAvailable(w, host)
	case NotFound:
		handleNotFound(w, host)
	default:
		handleInternalError(w, host)
	}
}

func handleNotFound(w http.ResponseWriter, host *Host) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, http.StatusText(http.StatusNotFound))
	Logger().Printf("[%s] host '%s' host does not exist\n", host.name, host.name)
}

func handleNotAvailable(w http.ResponseWriter, host *Host) {
	w.WriteHeader(http.StatusServiceUnavailable)
	fmt.Fprint(w, http.StatusText(http.StatusServiceUnavailable))
	Logger().Printf("[%s] host host is currently not available\n", host.name)
}

func handleInternalError(w http.ResponseWriter, host *Host) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
	Logger().Printf("[%s] unknown error occurred: %s\n", host.name, host.Error)
}
