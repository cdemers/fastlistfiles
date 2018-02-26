package main

import (
	_ "expvar"
	"flag"
	"fmt"
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
	var dirname = flag.String("folder", ".", "The folder to scan.")
	var sorted = flag.Bool("sort", false, "Sort files, defaults to unsorted.")
	var includeDirs = flag.Bool("dirs", false, "Include directories, default to not.")
	var expvarPort = flag.String("expvar-port", "", "The port number for the expvar instrumentation service.")
	var workercount = flag.Int("workers", 20, "The number of workers to use, it can go pretty high.")

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

	var jobsIdCounter int

	jobs := make(chan ListFilesJob, *workercount*10)
	defer close(jobs)

	jobResults := make(chan ListFilesResult, 100000000) // Woah! 100 millions FTW! :D
	defer close(jobResults)

	// Start the workers
	for workerId := 0; workerId < *workercount; workerId++ {
		go worker(workerId, jobs, jobResults)
	}

	// folders, err := splitFoldersTree(*dirname)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	//
	// for _, v := range folders {
	// 	fmt.Println(v)
	// }

	{
		jobs <- ListFilesJob{
			JobID:       jobsIdCounter,
			BasePath:    *dirname,
			IncludeDirs: *includeDirs,
			Sorted:      *sorted,
		}

		jobsIdCounter++
	}

	for a := 1; a <= jobsIdCounter; a++ {
		_ = <-jobResults
		// fmt.Printf("Result Worker %d job %d, Error: %v\n", r.WorkerID, r.JobID, r.Error)
	}

}

func worker(workerID int, jobs <-chan ListFilesJob, results chan<- ListFilesResult) {

	for job := range jobs {

		err := godirwalk.Walk(job.BasePath, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if de.IsDir() && !job.IncludeDirs {
					return nil
				}
				fmt.Printf("%s\n", osPathname)
				return nil
			},
			Unsorted: job.Sorted,
		})

		results <- ListFilesResult{
			WorkerID: workerID,
			Error:    err,
			JobID:    job.JobID,
		}

	}

}

func splitFoldersTree(basePath string) (paths []string, err error) {

	err = godirwalk.Walk(basePath, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				return nil
			}

			fmt.Printf("Folder: %s\n", osPathname)
			return nil
		},
		Unsorted: true,
	})

	return []string{"aaa", "bbb"}, err
}
