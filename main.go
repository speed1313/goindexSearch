package main

import (
	"flag"
	"fmt"

	"sync"

	"github.com/speed1313/goindexSearch/searcher"
)

func main() {
	// Make HTTP GET request
	enumCount := uint64(0)
	var wg sync.WaitGroup
	pkgLists := searcher.GetPkgList()
	// get search way from command
	searcher := flag.String("cmd", "", "way of search")
	flag.Parse()
	var s searcher.EnumSearcher
	switch *searcher {
	case "vet":
		s = searcher.VetSearcher{}
	case "grep":
		s = searcher.GrepSearcher{}
	default:
		s = searcher.VetSearcher{}
	}
	for _, pkgname := range pkgLists[0:10] {
		wg.Add(1)
		go func(pkgname string, enumCount *uint64) {
			defer func() {
				wg.Done()
			}()
			searcher.EnumSearch(pkgname, enumCount, s)
		}(pkgname, &enumCount)
	}
	wg.Wait()
	fmt.Printf("enum count: %d\n", enumCount)
}
