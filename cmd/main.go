package main

import (
	"fmq/internal/logger"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

var forceMaxQuality bool
var debug bool

var paramsToStrip = []string{
	"maxVideoBitrate",
	"videoResolution",
	"videoQuality",
}

func main() {
	forceMaxQuality = strings.ToLower(os.Getenv("FORCE_MAX_QUALITY")) == "true"
	debug = strings.ToLower(os.Getenv("DEBUG")) == "true"

	plexUrl := os.Getenv("PLEX_URL")
	if plexUrl == "" {
		logger.Fatal("Missing Plex URL from env vars")
	}

	target, err := url.Parse(plexUrl)
	if err != nil {
		logger.Fatal("Failed to parse target Plex URL: %v", err)
	}

	if err := testPlexConnection(target); err != nil {
		logger.Fatal("Failed to reach Plex: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director 
	proxy.Director = func(req *http.Request) {
		logger.Info("%s %s from %s", req.Method, req.URL.Path, req.RemoteAddr)
		if debug {
			logger.Debug("Full query: %s", req.URL.RawQuery)
		}

		originalDirector(req)
		if forceMaxQuality && (strings.HasPrefix(req.URL.Path, "/video") || strings.Contains(req.URL.RawQuery, "Playback")) {
			stripQualityParams(req)
		}
	}

	logger.Info("Proxy started on :80\nPLEX_URL=%s\nFORCE_MAX_QUALITY=%t\nDEBUG=%t", plexUrl, forceMaxQuality, debug)
	if err := http.ListenAndServe(":80", proxy); err != nil {
		log.Fatalf("Proxy failed: %v", err)
	}
}

func stripQualityParams(req *http.Request) {
	var strippedParams []string

	q := req.URL.Query()
	for _, param := range paramsToStrip {
		if q.Has(param) {
			q.Del(param)
			strippedParams = append(strippedParams, param)
		}
	}
	req.URL.RawQuery = q.Encode()
	if debug && len(strippedParams) > 0 {
		logger.Debug("Stripped %v from %s", strippedParams, req.URL.Path)
	}
}

func testPlexConnection(url *url.URL) error {
	client := &http.Client{ Timeout: 10 * time.Second }

	resp, err := client.Get(url.String())
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}