package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"armbian-stats/api"
)

func main() {
	cfgPath := "config.yml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := api.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("[config] %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Printf("[server] starting on http://%s  interval=%ds", addr, cfg.Interval)

	collector := api.NewCollector()
	collector.Collect()
	time.Sleep(time.Duration(cfg.Interval) * time.Second)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
        t, err := template.New("ui").Parse(api.Gui)
		if err != nil {
			http.Error(w, "template error: "+err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.Execute(w, cfg.Theme); err != nil {
			log.Printf("[web] template error: %v", err)
		}
	})

	mux.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		log.Printf("[sse] connected %s", r.RemoteAddr)
		defer log.Printf("[sse] disconnected %s", r.RemoteAddr)

		ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
		defer ticker.Stop()

		sendStats(w, flusher, collector)

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				sendStats(w, flusher, collector)
			}
		}
	})

	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(collector.Collect())
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"ok","ts":%d}`, time.Now().Unix())
	})

	fmt.Printf("armbian-stats  addr=%s  config=%s  interval=%ds\n", addr, cfgPath, cfg.Interval)

	server := &http.Server{
		Addr:        addr,
		Handler:     loggingMiddleware(mux),
		ReadTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[server] %v", err)
	}
}

func sendStats(w http.ResponseWriter, f http.Flusher, c *api.Collector) {
	data, err := json.Marshal(c.Collect())
	if err != nil {
		log.Printf("[sse] marshal error: %v", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	f.Flush()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/stream") {
			log.Printf("[http] %s %s", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}
