package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	apachelog "github.com/lestrrat-go/apache-logformat"
)

func main() {
	portPtr := flag.Int("port", 8000, "HTTP Port to Listen")
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
	addr := ":" + strconv.Itoa(*portPtr)
	fmt.Printf("Serving %v on http://localhost%v\n", dir, addr)
	http.ListenAndServe(addr, apachelog.CombinedLog.Wrap(http.FileServer(http.Dir(dir)), os.Stderr))
}
