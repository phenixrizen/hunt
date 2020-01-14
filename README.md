# hunt
A go program to quickly hunt for content in source files


```
hunt - a simple way to hunt for content in source files

usage: hunt -query "foo bar" -root .

flags:
  -c, --c-files              search for c/c++ files
  -g, --go-files             search for Go files
  -j, --js-files             search for JavaScript files
  -n, --name-regexp string   regexp to match the filename
  -q, --query string         regexp to match source content
  -r, --root string          root to start your hunt
  -b, --ruby-files           search for ruby files
  -s, --rust-files           search for rust files
```

### Examples

Hunt for Go and C source files that have "Accept-Encoding" or "User-Agent" in the code:
```bash
$ hunt --query "Accept-Encoding|User-Agent" --root . --go-files --c-files
```
