package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	apachelog "github.com/lestrrat-go/apache-logformat"
)

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default: // linux, freebsd, openbsd, netbsd
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

func main() {
	var port int
	var open bool
	flag.IntVar(&port, "port", 0, "HTTP Port to Listen (0 for any available port)")
	flag.BoolVar(&open, "open", false, "Open browser on started")
	flag.Parse()
	if flag.NArg() > 1 {
		fmt.Println("Usage: servedir /path/to/host")
		os.Exit(1)
	}
	dir := "."
	if flag.NArg() == 1 {
		dir = flag.Args()[0]
	}
	dir, _ = filepath.Abs(dir)
	addr := ":" + strconv.Itoa(port)
	
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating listener: %v\n", err)
		os.Exit(1)
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port
	url := "http://localhost:" + strconv.Itoa(actualPort)
	
	fmt.Printf("Serving %v on %v\n", dir, url)
	if open {
		go func() {
			time.Sleep(time.Second)
			_ = openBrowser(url)
		}()
	}
	if err := http.Serve(listener, apachelog.CombinedLog.Wrap(http.FileServer(http.Dir(dir)), os.Stderr)); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
