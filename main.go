package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/speed1313/goindexSearch/searcher"
)

func main() {
	enumCount := uint64(0)
	pkgCount := uint64(0)
	since := flag.String("since", "2019-04-10T19:08:52.997264Z", "since time in RFC3339 format")
	last := flag.String("last", "2019-05-10T19:08:52.997264Z", "last time in RFC3339 format")
	cmd := flag.String("cmd", "", "Way of search. vet(default) or grep")
	vettoolPath := flag.String("vettool", "", "Path of vet tool")
	pattern := flag.String("pattern", `\benum\b`, "pattern of grep")
	searchNum := flag.Int("n", 10, "number of packages to search")
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
	start := time.Now()
	sinceTime, err := time.Parse(time.RFC3339, *since)
	if err != nil {
		log.Fatal("parse since time failed: ", err)
	}
	lastTime, err := time.Parse(time.RFC3339, *last)
	if err != nil {
		log.Fatal("parse last time failed: ", err)
	}
	pkgLists := searcher.GetPkgList(sinceTime, lastTime)
	fmt.Printf("Number of packages between %s and %s: %d\n", sinceTime.Format(time.DateOnly), lastTime.Format(time.DateOnly), len(pkgLists))
	var wg sync.WaitGroup
	wg.Add(*searchNum)
	maxGoroutines := runtime.NumCPU()
	guard := make(chan struct{}, maxGoroutines)

	// speed up idea
	// - Execute go vet after filtering by grep.
	for _, pkgname := range pkgLists[:*searchNum] {
		guard <- struct{}{}
		go func(pkgname string) {
			defer func() {
				wg.Done()
				<-guard
			}()
			searcher.EnumSearch(pkgname, &enumCount, &pkgCount, s)
		}(pkgname)
	}
	wg.Wait()

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Number of packages which is pointed out: %d/%d [%s]\n", enumCount, pkgCount, elapsed)
}
