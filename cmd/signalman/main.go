package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"

	"jarnfast/signalman/pkg"
)

var version string = "1.0"
var build string = "n/a"

type handlers struct {
}

type adminPortal struct {
}

func newAdminPortal() *adminPortal {
	return &adminPortal{}
}

func (h *handlers) status(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(versionString()))
	response.Write([]byte(fmt.Sprintf("Args: %d, %s", len(os.Args[1:]), os.Args[1:])))
}

func (h *handlers) signal(response http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Get request", request.URL)
}
func (h *handlers) kill(response http.ResponseWriter, request *http.Request) {
	os.Exit(1337)
}
func versionString() string {
	//return fmt.Sprintf("%s version %s (build %s)", path.Base(os.Args[0]), version, build)
	return fmt.Sprintf("%s version %s (build %s)", path.Base(os.Args[0]), pkg.Version, build)
}

func main() {
	fmt.Println("Starting", versionString())

	addr := os.Getenv("SIGNALMAN_LISTEN_ADDRESS")
	if addr == "" {
		addr = "localhost:30000"
	}

	h := &handlers{}
	http.HandleFunc("/status", h.status)
	http.HandleFunc("/kill", h.kill)
	http.HandleFunc("/signal/", h.signal)

	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("Listening for commands on", addr)

	args := os.Args[1:]
	binary, err := exec.LookPath(args[0])
	if err != nil {
		fmt.Println("Command not found", binary)
		panic(err)
	}

	cmd := exec.Command(binary, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		fmt.Println("Unable to run command", err)
		panic(err)
	}

	cmd.Wait()

	fmt.Println("Command terminated")
}
