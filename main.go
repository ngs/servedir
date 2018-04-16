package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	apachelog "github.com/lestrrat-go/apache-logformat"
)

func main() {
	dir := os.Args[1:]
	if len(dir) != 1 {
		fmt.Println("Usage: servedir /path/to/host")
		os.Exit(1)
	}
	portPtr := flag.Int("port", 8000, "HTTP Port to Listen")
	addr := ":" + strconv.Itoa(*portPtr)
	fmt.Printf("HTTP Server started on http://localhost%s\n", addr)
	http.ListenAndServe(addr, apachelog.CombinedLog.Wrap(http.FileServer(http.Dir(dir[0])), os.Stderr))
}
