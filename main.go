package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/karrick/godirwalk"
)

type ListFilesJob struct {
	JobID       int
	BasePath    string
	IncludeDirs bool
	Sorted      bool
}

type ListFilesResult struct {
	WorkerID int
	JobID    int
	Error    error
}

func main() {
	// var isCompleted bool

	var basePath = flag.String("folder", ".", "The folder to scan.")
	var sorted = flag.Bool("sort", false, "Sort files, defaults to unsorted.")
	var includeDirs = flag.Bool("dirs", false, "Include directories, defaults to no.")
	var expvarPort = flag.String("expvar-port", "", "The port number for the expvar instrumentation service.")
	var includeHiddens = flag.Bool("include-hiddens", false, "Include hidden files and folders. False by default.")

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

	go worker(completionChannel, *basePath, *includeDirs, *sorted, *includeHiddens)

	_ = <-completionChannel

}

func worker(completed chan<- bool, basePath string, includeDirs, sorted, includeHiddens bool) {

	err := godirwalk.Walk(basePath, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !includeDirs && de.IsDir() {
				return nil
			}
			if !includeHiddens && osPathname[0:1] == "." {
				return nil
			}
			fmt.Printf("%s\n", osPathname)
			return nil
		},
		Unsorted: !sorted,
	})

	if err != nil {
		log.Fatalln(err)
	}

	completed <- true
}
