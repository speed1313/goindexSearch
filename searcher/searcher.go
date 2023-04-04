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
	"time"

	"golang.org/x/exp/slog"
)

// get golang package list between since to last from "https://index.golang.org/index"
// the time format is RFC3339
func GetPkgList(since, last time.Time) []string {
	nextSince := since
	pkgLists := make([]string, 0)
	type Message struct {
		Path, Version, Timestamp string
	}
	url := "https://index.golang.org/index?since="
	for {
		response, err := http.Get(url + nextSince.Format(time.RFC3339))
		if err != nil {
			log.Fatal("get package list failed: ", err)
		}
		defer response.Body.Close()
		scanner := bufio.NewScanner(response.Body)
		var m Message
		for scanner.Scan() {
			if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {

				log.Fatal("unmarshal failed: ", err)
			}
			pkgLists = append(pkgLists, m.Path)
		}
		pkgLists = removeDuplicate(pkgLists)
		nextSince, err = time.Parse(time.RFC3339, m.Timestamp)
		if err != nil {
			log.Fatal("parse time failed: ", err)
		}
		if nextSince.After(last) {
			break
		}
	}
	pkgLists = removeDuplicate(pkgLists)
	return pkgLists
}

type EnumSearcher interface {
	Search(dir string, pkgname string, ch chan<- string, pkgch chan<- string) error
}

func EnumSearch(pkgname string, searcher EnumSearcher, ch chan<- string, pkgch chan<- string) error {
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
	if err := searcher.Search(dir, pkgname, ch, pkgch); err != nil {
		return err
	}
	return nil
}

// cleanWorkSpace clean tmp directories and go clean -i packages
func cleanWorkSpace(pkgname, dir string) error {
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

type VetSearcher struct {
	VettoolPath string
}

func (v VetSearcher) Search(dir string, pkgname string, ch chan<- string, pkgch chan<- string) error {
	arg := path.Join(pkgname, "...")
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		return err
	} else {
		pkgch <- pkgname
	}
	option := []string{"vet"}

	if v.VettoolPath != "" {
		option = append(option, "-vettool", v.VettoolPath)
	}
	option = append(option, arg)
	cmd = exec.Command("go", option...)
	cmd.Dir = path.Join(".", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// For now, go vet any error is counted.
		ch <- pkgname
		slog.Debug("vet", "out", string(out))
		fmt.Println(pkgname)
	}
	return nil
}

type GrepSearcher struct {
	Pattern string
}

// "grep  $(go list -f '{{.Dir}}' $(go list -f '{{join .Deps "\n"}}' a))"
func (g GrepSearcher) Search(dir string, pkgname string, ch chan<- string, pkgch chan<- string) error {
	arg := path.Join(pkgname, "...")
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		return err
	} else {
		pkgch <- pkgname
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
			slog.Debug("grep", "out", string(out))
			break
		}
	}
	if isEnumUsed {
		fmt.Println(pkgname)
		ch <- pkgname
	}

	return nil
}
