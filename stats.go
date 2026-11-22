package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/nsgo-dev/nsgo-pet/pet"
	"github.com/nsgo-dev/nsgo/core"
)

// The app counts its use in the website's Umami, as pages under /app/ of
// the site's host: a pageview as it starts and whenever the user chooses
// another pet, tagged "app" as the site's own pages are "website". An
// install carries a random ID kept in the storage, so that one install is
// one visitor. Only builds of nsgo build report: nsgo dev, nsgo test and
// plain go builds send nothing.
const (
	statsURL     = "https://umami.oneo.me/api/send"
	statsWebsite = "28bf766f-61e5-4b0a-9dc3-9b5d2d33c8a4"
	statsHost    = "nsgo-dev.github.io"
	statsPath    = "/app/"
	statsTag     = "app"
)

// installKey is the storage key of the install's ID.
const installKey = "installID"

// statsPayload is a pageview of Umami's /api/send.
type statsPayload struct {
	Website  string `json:"website"`
	Hostname string `json:"hostname"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Language string `json:"language,omitempty"`
	Tag      string `json:"tag"`
	ID       string `json:"id"`
}

// statsReporter returns the function that counts a pet shown, or one that
// does nothing outside release builds. Pageviews are sent from another
// goroutine, oldest first; one that fails is dropped.
func statsReporter(cx *core.Scope) func(p *pet.Pack) {
	if core.App().Development {
		return func(*pet.Pack) {}
	}
	env := cx.Env()
	id := installID(cx)
	queue := make(chan statsPayload, 8)
	go sendStats(queue, statsAgent(env.OS()))
	return func(p *pet.Pack) {
		pv := statsPayload{Website: statsWebsite, Hostname: statsHost, URL: statsPath + neturl.PathEscape(p.ID),
			Title: p.Name, Language: env.Locale(), Tag: statsTag, ID: id}
		select {
		case queue <- pv:
		default:
		}
	}
}

// installID returns the install's random ID, made on first use.
func installID(cx *core.Scope) string {
	store := cx.Storage()
	id, ok := store.GetItem(installKey)
	if !ok {
		b := make([]byte, 16)
		rand.Read(b)
		id = hex.EncodeToString(b)
		store.SetItem(installKey, id)
	}
	return id
}

// sendStats posts the queue's pageviews to Umami until it closes.
func sendStats(queue <-chan statsPayload, agent string) {
	client := &http.Client{Timeout: 10 * time.Second}
	for p := range queue {
		body, _ := json.Marshal(map[string]any{"type": "event", "payload": p})
		req, err := http.NewRequest(http.MethodPost, statsURL, bytes.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if agent != "" {
			req.Header.Set("User-Agent", agent)
		}
		if resp, err := client.Do(req); err == nil {
			resp.Body.Close()
		}
	}
}

// statsAgent is the User-Agent the app sends, as a browser on the platform
// would say it, for Umami to tell the system and not take it for a bot as
// it does Go's own.
func statsAgent(os core.OS) string {
	platforms := map[core.OS]string{
		core.OSMacOS:   "Macintosh; Intel Mac OS X 10_15_7",
		core.OSWindows: "Windows NT 10.0; Win64; x64",
		core.OSLinux:   "X11; Linux x86_64",
	}
	p, ok := platforms[os]
	if !ok {
		return ""
	}
	return "Mozilla/5.0 (" + p + ") nsgo-pet/" + core.App().Version
}
