# goindexSearch
goindexsearch is a package that can do go vet and grep for all packages in https://index.golang.org/index

# Features
- You can use grep or custom vet tool.

# How to use
- Install vettool you want to use.
```
$ go get github.com/speed1313/enumResearch/cmd/enumResearch
$ go install github.com/speed1313/enumResearch/cmd/enumResearch
```

- Run goindexSearch
```
$ go run main.go -cmd vet -vettool /path/of/enumResearch
golang.org/x/sys is using enum
golang.org/x/crypto is using enum
golang.org/x/text is using enum
Number of packages which use enum: 3/6 [30.459553879s]
```

- Run goindexSearch with grep
```
$ go run main.go -cmd grep  -pattern "\benum\b"
golang.org/x/sys is using enum
golang.org/x/text is using enum
Number of packages which use enum: 2/6 [13.4569623s]
```