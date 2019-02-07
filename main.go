package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/karrick/godirwalk"
)

func main() {

	var basepath = flag.String("folder", ".", "The folder to scan.")
	var sorted = flag.Bool("sort", false, "Sort files, defaults to unsorted.")
	var includeDirs = flag.Bool("dirs", false, "Include directories, defaults to no.")
	var expvarPort = flag.String("expvar-port", "", "The port number for the expvar instrumentation service.")
	var includeHiddens = flag.Bool("include-hiddens", false, "(DISABLED in v0.6.3, hidden files and folders will always be ignored) Include hidden files and folders. False by default.")

	flag.Parse()

	if *expvarPort != "" {
		sock, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", *expvarPort))
		if err != nil {
			fmt.Printf("Unable to serve on port %s.\nError: \"%v\"\n", *expvarPort, err)
			os.Exit(1)
		}
		go func() {
			fmt.Printf("HTTP expvar instrumentation metrics server now available at port %s\n", *expvarPort)
			fmt.Printf("You could monitor this program using the following command: \"expvarmon -ports %s\"\n", *expvarPort)
			http.Serve(sock, nil)
		}()
	}

	completionChannel := make(chan bool)

	go worker(completionChannel, *basepath, *includeDirs, *sorted, *includeHiddens)

	_ = <-completionChannel

}

func worker(completed chan<- bool, basepath string, includeDirs, sorted, includeHiddens bool) {

	err := godirwalk.Walk(basepath, &godirwalk.Options{
		Unsorted:      !sorted,
		IgnoreHiddens: !includeHiddens,
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !includeDirs && de.IsDir() {
				return nil
			}
			fmt.Printf("%s\n", osPathname)
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			if err.Error() == "Callback: SKIPx" {
				return godirwalk.SkipNode
			}
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return godirwalk.SkipNode
		},
	})

	if err != nil {
		log.Fatalln(err)
	}

	completed <- true
}
