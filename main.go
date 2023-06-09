package main

import (
	"flag"
	"fmt"
	"log"

	"os"
	"runtime"
	"sync"
	"time"

	"github.com/speed1313/goindexSearch/searcher"
	"golang.org/x/exp/slog"
)

type config struct {
	Since        time.Time
	Last         time.Time
	Cmd          string
	VettoolPath  string
	Pattern      string
	SearchNumber int
	Sercher      searcher.EnumSearcher
}

const (
	defaultSince = "2019-04-10T19:08:52.997264Z"
	defaultLast  = "2019-04-10T19:08:52.997264Z"
)

func newConfig() *config {
	since := flag.String("since", defaultSince, "since time in RFC3339 format")
	last := flag.String("last", defaultLast, "last time in RFC3339 format")
	cmd := flag.String("cmd", "", "Way of search. vet(default) or grep")
	vettoolPath := flag.String("vettool", "", "Path of vet tool")
	pattern := flag.String("pattern", `\benum\b`, "pattern of grep")
	searchNum := flag.Int("n", 10, "number of packages to search")
	isVerbose := flag.Bool("v", false, "verbose output")
	flag.Parse()
	var s searcher.EnumSearcher
	switch *cmd {
	case "vet":
		s = searcher.VetSearcher{VettoolPath: *vettoolPath}
	case "grep":
		s = searcher.GrepSearcher{Pattern: *pattern}
	default:
		log.Fatal("choose vet or grep")
	}
	sinceTime, err := time.Parse(time.RFC3339, *since)
	if err != nil {
		log.Fatal("parse since time failed: ", err)
	}
	lastTime, err := time.Parse(time.RFC3339, *last)
	if err != nil {
		log.Fatal("parse last time failed: ", err)
	}
	if *isVerbose {
		var programLevel = new(slog.LevelVar) // Info by default
		h := slog.HandlerOptions{Level: programLevel}.NewTextHandler(os.Stderr)
		slog.SetDefault(slog.New(h))
		programLevel.Set(slog.LevelDebug)
	}
	return &config{
		Since:        sinceTime,
		Last:         lastTime,
		Cmd:          *cmd,
		VettoolPath:  *vettoolPath,
		Pattern:      *pattern,
		SearchNumber: *searchNum,
		Sercher:      s,
	}
}

func main() {
	c := newConfig()

	start := time.Now()

	pkgLists := searcher.GetPkgList(c.Since, c.Last)
	fmt.Printf("Number of packages between %s and %s: %d\n", c.Since.Format(time.DateOnly), c.Last.Format(time.DateOnly), len(pkgLists))
	var wg sync.WaitGroup
	wg.Add(c.SearchNumber)
	maxGoroutines := runtime.NumCPU()
	guard := make(chan struct{}, maxGoroutines)
	ch := make(chan string, maxGoroutines)
	pkgch := make(chan string, maxGoroutines)
	searchedPackageList := make([]string, 0)
	go func() {
		for r := range ch {
			searchedPackageList = append(searchedPackageList, r)
		}
	}()
	targetPackageList := make([]string, 0)
	go func() {
		for r := range pkgch {
			targetPackageList = append(targetPackageList, r)
		}
	}()
	// speed up idea
	// - Execute go vet after filtering by grep.
	for _, pkgname := range pkgLists[:c.SearchNumber] {
		guard <- struct{}{}
		go func(pkgname string, ch chan<- string, pkgch chan<- string) {
			defer func() {
				wg.Done()
				<-guard
			}()
			searcher.EnumSearch(pkgname, c.Sercher, ch, pkgch)
		}(pkgname, ch, pkgch)
	}
	wg.Wait()
	close(ch)
	close(pkgch)
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Number of packages which is pointed out: %d/%d [%s]\n", len(searchedPackageList), len(targetPackageList), elapsed)
}
