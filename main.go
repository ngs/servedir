package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	apachelog "github.com/lestrrat-go/apache-logformat"
)

func main() {
	var port int
	var open bool
	flag.IntVar(&port, "port", 8000, "HTTP Port to Listen")
	if runtime.GOOS == "darwin" {
		flag.BoolVar(&open, "open", false, "Open browser on started")
	}
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
	url := "http://localhost" + addr
	fmt.Printf("Serving %v on %v\n", dir, url)
	if open {
		go func() {
			time.Sleep(time.Second)
			exec.Command("/usr/bin/open", url).Run()
		}()
	}
	http.ListenAndServe(addr, apachelog.CombinedLog.Wrap(http.FileServer(http.Dir(dir)), os.Stderr))
}
