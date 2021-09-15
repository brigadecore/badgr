package main

import (
	"log"
	"net/http"

	"github.com/brigadecore/badgr/internal/badges"
	"github.com/brigadecore/badgr/internal/badges/redis"
	libHTTP "github.com/brigadecore/brigade-foundations/http"
	"github.com/brigadecore/brigade-foundations/signals"
	"github.com/brigadecore/brigade-foundations/version"
	"github.com/gorilla/mux"
)

func main() {
	log.Printf(
		"Starting Badgr -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	cacheConfig, err := redisCacheConfig()
	if err != nil {
		log.Fatal(err)
	}

	handler := badges.NewHandler(
		badges.NewService(),
		redis.NewCache(cacheConfig),
	)

	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc(
		"/v1/github/checks/{owner}/{repo}/badge.svg",
		handler.ServeHTTP,
	).Methods(http.MethodGet)
	router.HandleFunc("/healthz", libHTTP.Healthz).Methods(http.MethodGet)

	serverConfig, err := serverConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(
		libHTTP.NewServer(
			router,
			&serverConfig,
		).ListenAndServe(signals.Context()),
	)

}
