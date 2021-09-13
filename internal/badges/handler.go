package badges

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

// handler is an implementation of the http.handler interface that can serve
// badges by by delegating to a transport-agnostic Service interface.
type handler struct {
	service Service
	cache   Cache
}

// NewHandler returns an implementation of the http.handler interface that can
// serve badges by by delegating to a transport-agnostic Service interface.
func NewHandler(service Service, cache Cache) http.Handler {
	return &handler{
		service: service,
		cache:   cache,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Search the warm cache
	if url, err := h.cache.GetWarm(r.URL.String()); err != nil {
		log.Printf(
			"error retrieving result for key %q from warm cache: %s",
			r.URL.String(),
			err,
		)
		// Don't return yet. We can still ask the service for a fresh result.
	} else if url != "" { // Warm cache hit!
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	// If we get to here, either the warm cache lookup failed or we had a warm
	// cache miss. Either way we'll ask the service for a fresh result.
	appIDStr := r.URL.Query().Get("appID")
	var appID int
	if appIDStr != "" {
		var err error
		appID, err = strconv.Atoi(appIDStr)
		if err != nil {
			http.Redirect(
				w,
				r,
				badgeURL(NewErrBadge(http.StatusBadRequest)),
				http.StatusSeeOther,
			)
			return
		}
	}
	if badge, err := h.service.CheckBadge(
		r.Context(),
		mux.Vars(r)["owner"],
		mux.Vars(r)["repo"],
		&CheckBadgeOptions{
			BadgeName:   r.URL.Query().Get("name"),
			GitHubAppID: appID,
			Branch:      r.URL.Query().Get("branch"),
		},
	); err != nil {
		log.Printf("error getting check badge: %s", err)
		// Don't return yet. We can still check the cold cache.
	} else { // A fresh badge
		// Try to cache this
		badgeURL := badgeURL(badge)
		if err := h.cache.Set(r.URL.String(), badgeURL); err != nil {
			log.Printf(
				"error writing result for key %q to cache: %s",
				r.URL.String(),
				err,
			)
		}
		http.Redirect(w, r, badgeURL, http.StatusSeeOther)
		return
	}

	// If we get to here, we didn't get anything from the warm cache and the
	// service errorred. Try the cold cache.
	if url, err := h.cache.GetCold(r.URL.String()); err != nil {
		log.Printf(
			"error retrieving result for key %q from cold cache: %s",
			r.URL.String(),
			err,
		)
	} else if url != "" { // Cold cache hit!
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	// If we get to here, we have been completely unsuccessful.
	http.Redirect(
		w,
		r,
		badgeURL(NewErrBadge(http.StatusInternalServerError)),
		http.StatusSeeOther,
	)
}

func badgeURL(badge Badge) string {
	return fmt.Sprintf(
		"https://img.shields.io/static/v1?label=%s&message=%s&color=%s",
		url.PathEscape(badge.Name()),
		url.PathEscape(badge.Status()),
		url.PathEscape(string(badge.Color())),
	)
}
