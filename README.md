# goindexSearch
goindexsearch is a package that can do go vet and grep for all packages in https://index.golang.org/index

# Features
- You can use grep or custom vet tool.
- You can search for all packages in https://index.golang.org/index in chronological order.

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
$  go run main.go -h
Usage of /var/folders/yc/fsbnfh950jbfby21gtfxgxyr0000gn/T/go-build1681217861/b001/exe/main:
  -cmd string
        Way of search. vet(default) or grep
  -last string
        last time in RFC3339 format (default "2019-04-10T19:08:52.997264Z")
  -n int
        number of packages to search (default 10)
  -pattern string
        pattern of grep (default "\\benum\\b")
  -since string
        since time in RFC3339 format (default "2019-04-10T19:08:52.997264Z")
  -vettool string
        Path of vet tool

$ go run main.go -cmd vet -vettool /Users/sugiurahajime/go/bin/enumResearch
Number of packages between 2019-04-10 and 2019-04-10: 864
golang.org/x/text
Number of packages which is pointed out: 1/6 [17.993530617s]

```

- Run goindexSearch with grep
```
$ go run main.go -cmd grep  -pattern "\benum\b" -n 30
Number of packages between 2019-04-10 and 2019-04-10: 864
golang.org/x/sys
golang.org/x/text
gocloud.dev
cloud.google.com/go
google.golang.org/api
Number of packages which is pointed out: 5/26 [33.128365943s]

```