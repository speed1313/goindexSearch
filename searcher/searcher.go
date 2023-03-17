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
	Search(dir string, pkgname string, enumCount *uint64) error
}

func EnumSearch(pkgname string, enumCount *uint64, searcher EnumSearcher) {
	dir := getHashedDir(pkgname)
	if _, err := os.Stat(dir); os.IsExist(err) {
		fmt.Printf("dir %s already exist", dir)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("mkdir %s failed: %s", pkgname, err)
	}
	// create go.mod
	if _, err := os.Stat(path.Join(dir, "go.mod")); os.IsNotExist(err) {
		cmd := exec.Command("go", "mod", "init", "a")
		cmd.Dir = path.Join(".", dir)
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s go mod init failed: %s\n", pkgname, err)
		}
	}
	if err := searcher.Search(dir, pkgname, enumCount); err != nil {
		fmt.Printf("enum search failed: %s\n", err)
	}
	cleanWorkSpace(pkgname, dir)
}

func cleanWorkSpace(pkgname, dir string) {
	arg := path.Join(pkgname, "...")
	// clean pkg
	cmd := exec.Command("go", "clean", "-i", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("go clean %s failed: %s\n", pkgname, err)
	}
	// remove dir
	if err := os.RemoveAll(dir); err != nil {
		log.Printf("remove dir %s failed: %s\n", dir, err)
	}
}

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

func getHashedDir(pkgname string) string {
	hashDir := sha256.Sum256([]byte(pkgname))
	dir := fmt.Sprintf("%x", hashDir[:8])
	dir = path.Join(".", "tmpdir", dir)
	return dir
}

type VetSearcher struct{}

func (VetSearcher) Search(dir string, pkgname string, enumCount *uint64) error {
	arg := path.Join(pkgname, "...")
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("go get %s failed: %s\n", pkgname, err)
		return err
	}
	cmd = exec.Command("go", "vet", "-vettool=/Users/sugiurahajime/go/bin/enumResearch", arg)
	cmd.Dir = path.Join(".", dir)
	_, err := cmd.CombinedOutput()
	if err != nil {
		// if -v is set, print output
		// print("vet output: ", string(out))
		atomic.AddUint64(enumCount, 1)
		println(pkgname, "is using enum")
	}
	return nil
}

type GrepSearcher struct{}

// "grep  $(go list -f '{{.Dir}}' $(go list -f '{{join .Deps "\n"}}' a))"
func (GrepSearcher) Search(dir string, pkgname string, enumCount *uint64) error {
	arg := path.Join(pkgname, "...")
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("go get %s failed: %s\n", pkgname, err)
		return err
	}
	cmd = exec.Command("go", "list", "-f", "{{.Dir}}", arg)
	cmd.Dir = path.Join(".", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("go list error", err)
	}
	isEnumUsed := false
	for _, targetdir := range strings.Split(string(out), "\n") {
		cmd = exec.Command("grep", "-r", `\benum\b`, targetdir)
		out, _ := cmd.Output()
		if len(out) > 0 {
			isEnumUsed = true
			// TODO: if -v is set, print out
			//fmt.Print(string(out))
		}
	}
	if isEnumUsed {
		fmt.Println(pkgname, "is using enum")
		atomic.AddUint64(enumCount, 1)
	}

	return nil
}
