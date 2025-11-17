package main

import (
	crand "crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

type syntheticLoggingRequest struct {
	Duration string `json:"duration"`
	Rate     int64  `json:"rate"`
}

type loggingQueue struct {
	end  <-chan time.Time
	rate int64
}

var workerCh chan loggingQueue

func main() {
	rand.New(rand.NewSource((time.Now().UnixNano())))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// spawn worker
	workerCh = make(chan loggingQueue, 5)
	go worker(workerCh)

	// router setup
	mux := http.NewServeMux()
	mux.HandleFunc("GET /timeout", timeout)
	mux.HandleFunc("GET /slow", slowHandler)
	mux.HandleFunc("GET /500", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("GET /", rootHandler)
	mux.HandleFunc("POST /", postRoot)
	router := loggingMiddleware(mux)
	logger.Info("Listening on :4000...")
	err := http.ListenAndServe(":4000", router)
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}

func worker(ch <-chan loggingQueue) {
	for lq := range ch {
		wait := int64(time.Second) / lq.rate
		go func() {
			for {
				select {
				case <-lq.end:
					return
				default:
					b := make([]byte, 10)
					_, err := crand.Read(b)
					if err != nil {
						slog.Error("rand failed", slog.String("err", err.Error()))
						return
					}
					slog.Info("some random fake data for log testing",
						"id", b,
						"extra", dummyResponse,
					)
					time.Sleep(time.Duration(wait))
				}
			}
		}()
	}
}

func postRoot(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var jb syntheticLoggingRequest
	if err = json.Unmarshal(body, &jb); err != nil {
		slog.Error("failed to unmarshal", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid json"))
		return
	}

	dur, err := time.ParseDuration(jb.Duration)
	if err != nil {
		slog.Error("failed to unmarshal", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid duration format"))
		return
	}
	lq := loggingQueue{
		end:  time.After(dur),
		rate: jb.Rate,
	}
	workerCh <- lq
}

func wrappedWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, []byte{}, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	lrw.body = data
	return lrw.ResponseWriter.Write(data)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			wrapped := wrappedWriter(w)
			next.ServeHTTP(wrapped, r)
			var level slog.Level
			if wrapped.statusCode >= 500 {
				level = slog.LevelError
			} else {
				level = slog.LevelInfo
			}
			slog.Log(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("host", r.Host),
				slog.String("path", r.URL.Path),
				slog.String("remoteaddr", r.RemoteAddr),
				slog.String("proto", r.Proto),
				slog.Int("code", wrapped.statusCode),
				// slog.String("body", string(wrapped.body)),
			)
		},
	)
}

func artificalDelay(min, max int) {
	n := rand.Intn(max-min+1) + min
	time.Sleep(time.Millisecond * time.Duration(n))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// artificalDelay(30, 70)
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(dummyResponse)
	if err != nil {
		slog.Error("marshal failed", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(data); err != nil {
		slog.Error("write failed", slog.String("err", err.Error()))
	}
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	artificalDelay(400, 500)
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(dummyResponse)
	if err != nil {
		slog.Error("marshal failed", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(data); err != nil {
		slog.Error("write failed", slog.String("err", err.Error()))
	}
}

func timeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 10)
	w.WriteHeader(http.StatusRequestTimeout)
}
