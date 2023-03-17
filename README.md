# goindexSearch
goindexsearch is a package that can do go vet and grep for all packages in https://index.golang.org/index

# Features
- You can use grep or custom vet tool.
- You can search for all packages in https://index.golang.org/index in chronological order. (Currently only the oldest 2000 packages can be searched.)

# Use cases
When adding new features like enum type to Golang, you may want to investigate a large number of packages to ensure backward compatibility.

# How to use
- Install vettool you want to use.
```
$ go get github.com/speed1313/enumResearch/cmd/enumResearch
$ go install github.com/speed1313/enumResearch/cmd/enumResearch
```

- Run goindexSearch with go vet
```
$ go run main.go -cmd vet -vettool /Users/sugiurahajime/go/bin/enumResearch
golang.org/x/text
Number of packages which is pointed out: 1/6 [21.194701473s]
```

- Run goindexSearch with grep
```
$ go run main.go -cmd grep  -pattern "\benum\b" -n 10
golang.org/x/sys
golang.org/x/text
Number of packages which is pointed out: 2/6 [7.064072991s]
```