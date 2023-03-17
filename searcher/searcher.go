package searcher
import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync/atomic"
)

// get package list from https://index.golang.org/index
func GetPkgList() []string {
	response, err := http.Get("https://index.golang.org/index")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	type Message struct {
		Path, Version, Timestamp string
	}
	pkgLists := make([]string, 0)
	for scanner.Scan() {
		var m Message
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			log.Fatal(err)
		}
		pkgLists = append(pkgLists, m.Path)
	}
	pkgLists = removeDuplicate(pkgLists)
	return pkgLists
}


type EnumSearcher interface {
	Search(dir string, pkgname string, enumCount *uint64, pkgCount *uint64) error
}

func EnumSearch(pkgname string, enumCount *uint64, pkgCount *uint64, searcher EnumSearcher) error{
	dir := getHashedDir(pkgname)
	defer cleanWorkSpace(pkgname, dir)

	if _, err := os.Stat(dir); os.IsExist(err) {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("mkdir %s failed: %s", pkgname, err)
	}
	// create go.mod
	if _, err := os.Stat(path.Join(dir, "go.mod")); os.IsNotExist(err) {
		cmd := exec.Command("go", "mod", "init", "a")
		cmd.Dir = path.Join(".", dir)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	if err := searcher.Search(dir, pkgname, enumCount, pkgCount); err != nil {
		return err
	}
	return nil
}

// cleanWorkSpace clean tmp directories and go clean -i packages
func cleanWorkSpace(pkgname, dir string) error{
	arg := path.Join(pkgname, "...")
	// clean pkg
	cmd := exec.Command("go", "clean", "-i", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		return err
	}
	// remove dir
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return nil
}

// removeDuplicate remove duplicate elements in slice
func removeDuplicate[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// getHashedDir get sha256 hash of the package name
func getHashedDir(pkgname string) string {
	hashDir := sha256.Sum256([]byte(pkgname))
	dir := fmt.Sprintf("%x", hashDir[:8])
	dir = path.Join(".", "tmpdir", dir)
	return dir
}

type VetSearcher struct{
	VettoolPath string
}

func (v VetSearcher) Search(dir string, pkgname string, enumCount *uint64, pkgCount *uint64) error {
	arg := path.Join(pkgname, "...")
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		return err
	}else{
		atomic.AddUint64(pkgCount, 1)
	}
	option := []string{"vet"}

	if v.VettoolPath != "" {
		option = append(option, "-vettool", v.VettoolPath)
	}
	option = append(option, arg)
	cmd = exec.Command("go", option...)
	cmd.Dir = path.Join(".", dir)
	_, err := cmd.CombinedOutput()
	if err != nil {
		// TODO: if -v is set, print output
		// print("vet output: ", string(out))
		atomic.AddUint64(enumCount, 1)
		println(pkgname)
	}
	return nil
}

type GrepSearcher struct{
	Pattern string
}

// "grep  $(go list -f '{{.Dir}}' $(go list -f '{{join .Deps "\n"}}' a))"
func (g GrepSearcher) Search(dir string, pkgname string, enumCount *uint64, pkgCount *uint64) error {
	arg := path.Join(pkgname, "...")
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		return err
	}else{
		atomic.AddUint64(pkgCount, 1)
	}
	cmd = exec.Command("go", "list", "-f", "{{.Dir}}", arg)
	cmd.Dir = path.Join(".", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("go list error", err)
	}
	isEnumUsed := false
	for _, targetdir := range strings.Split(string(out), "\n") {
		cmd = exec.Command("grep", "-r", g.Pattern, targetdir)
		out, _ := cmd.Output()
		if len(out) > 0 {
			isEnumUsed = true
			// TODO: if -v is set, print out
			//fmt.Print(string(out))
		}
	}
	if isEnumUsed {
		fmt.Println(pkgname)
		atomic.AddUint64(enumCount, 1)
	}

	return nil
}
