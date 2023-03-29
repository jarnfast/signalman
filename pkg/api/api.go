package api

import (
	"fmt"
	"jarnfast/signalman/pkg"
	"jarnfast/signalman/pkg/utl"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"net/http"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

type Api struct {
	addr   string
	logger *zap.SugaredLogger
	sigs   chan<- os.Signal
}

func NewApi(logger *zap.SugaredLogger) *Api {
	addr := utl.GetenvDefault("SIGNALMAN_LISTEN_ADDRESS", "localhost:30000")

	return &Api{
		addr:   addr,
		logger: logger,
	}
}

// Notify causes api to relay signal received on HTTP to c.
func (a *Api) Notify(c chan<- os.Signal) {
	a.sigs = c
}

func (a *Api) debugLogRequest(r *http.Request) {
	a.logger.Debugf("HTTP Request: %s %s", r.Method, r.URL)
}

func (a *Api) handleStatus(w http.ResponseWriter, r *http.Request) {
	a.debugLogRequest(r)

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pkg.VersionString()))
}

func (a *Api) handleSignal(w http.ResponseWriter, r *http.Request) {
	a.debugLogRequest(r)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	i, err := strconv.Atoi(parts[2])
	if err != nil {
		m := fmt.Sprintf("Unable to parse signal as int: %s", parts[2])
		a.logger.Debug(m)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(m))
		return
	}

	w.WriteHeader(http.StatusAccepted)

	s := syscall.Signal(i)

	a.sigs <- s

	w.Write([]byte(fmt.Sprintf("Got signal (%d) %s\r\n", len(parts), parts[2])))
	w.Write([]byte("This signal is: "))
	w.Write([]byte(strconv.Itoa(i)))
}

func (a *Api) handleTerm(w http.ResponseWriter, r *http.Request) {
	a.debugLogRequest(r)

	w.WriteHeader(http.StatusAccepted)

	a.sigs <- syscall.SIGTERM

	timeval := cast.ToDuration(utl.GetenvDefault("SIGNALMAN_TERM_TIMEOUT", "10"))
	<-time.After(timeval * time.Second)

	a.logger.Warnf("Wrapped command did not react to TERM within %d seconds. Sending KILL.", timeval)

	a.sigs <- syscall.SIGKILL
}

func (a *Api) handleKill(w http.ResponseWriter, r *http.Request) {
	a.debugLogRequest(r)

	w.WriteHeader(http.StatusAccepted)
	a.sigs <- syscall.SIGKILL
}

// ListenAndServe listens on the configured TCP network address and relays received
// signals to the configured channel
func (a *Api) ListenAndServe() {

	if a.sigs == nil {
		a.logger.Panic("No signal channel found")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/status", a.handleStatus)
	mux.HandleFunc("/signal/", a.handleSignal)
	mux.HandleFunc("/term", a.handleTerm)
	mux.HandleFunc("/kill", a.handleKill)

	go func() {
		err := http.ListenAndServe(a.addr, mux)
		if err != nil {
			a.logger.Panic(err)
		}
	}()

	a.logger.Infof("Listening for commands on %s", a.addr)
}
