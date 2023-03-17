package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/speed1313/goindexSearch/searcher"
)

func main() {
	// Make HTTP GET request
	enumCount := uint64(0)
	pkgCount := uint64(0)
	pkgLists := searcher.GetPkgList()
	// get search way from command
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
	var wg sync.WaitGroup
	wg.Add(*searchNum)
	maxGoroutines := 10
	guard := make(chan struct{}, maxGoroutines)
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
