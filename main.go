package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/spf13/pflag"
)

var root, query, filenameRegEx string
var cFiles, goFiles, rustFiles, rubyFiles, jsFiles bool
var qExpr *regexp.Regexp
var fExprs = make([]*regexp.Regexp, 0)
var found = 0
var wg sync.WaitGroup

func readFile(wg *sync.WaitGroup, path string) {
	defer wg.Done()

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		txt := scanner.Text()
		if qExpr.Match([]byte(txt)) {
			found++
			blue := color.New(color.FgBlue).SprintFunc()
			white := color.New(color.FgWhite).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			fmt.Printf("%s:%s    %s\n", blue(path), red(fmt.Sprintf("%d", i)), white(strings.TrimSpace(txt)))

		}
	}
}

func compExp(exp string) (*regexp.Regexp, error) {
	expr, err := regexp.Compile(exp)
	if err != nil {
		return nil, fmt.Errorf("error compiling file name regexp: %s", err)
	}
	return expr, nil
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func main() {
	pflag.StringVarP(&query, "query", "q", "", "regexp to match source content")
	pflag.StringVarP(&root, "root", "r", "", "root to start your hunt")
	pflag.StringVarP(&filenameRegEx, "name-regexp", "n", "", "regexp to match the filename")
	pflag.BoolVarP(&cFiles, "c-files", "c", false, "search for c/c++ files")
	pflag.BoolVarP(&goFiles, "go-files", "g", false, "search for Go files")
	pflag.BoolVarP(&rustFiles, "rust-files", "s", false, "search for rust files")
	pflag.BoolVarP(&rubyFiles, "ruby-files", "b", false, "search for ruby files")
	pflag.BoolVarP(&jsFiles, "js-files", "j", false, "search for JavaScript files")
	pflag.Parse()

	if pflag.NFlag() == 0 || root == "" {
		fmt.Println("hunt - a simple way to hunt for content in source files\n")
		fmt.Println("usage: hunt -query \"foo bar\" -root .\n")
		fmt.Println("flags:")
		pflag.PrintDefaults()
		os.Exit(-1)
	}

	expr, err := compExp(query)
	checkErr(err)
	qExpr = expr

	if filenameRegEx != "" {
		expr, err := compExp(filenameRegEx)
		checkErr(err)
		fExprs = append(fExprs, expr)
	}

	if cFiles || goFiles || rustFiles || rubyFiles || jsFiles {
		if cFiles {
			expr, err := compExp(".c$|.h$|.cpp$|.hpp$")
			checkErr(err)
			fExprs = append(fExprs, expr)
		}
		if goFiles {
			expr, err := compExp(".go$")
			checkErr(err)
			fExprs = append(fExprs, expr)
		}
		if rustFiles {
			expr, err := compExp(".rs$")
			checkErr(err)
			fExprs = append(fExprs, expr)
		}
		if rubyFiles {
			expr, err := compExp(".rb$")
			checkErr(err)
			fExprs = append(fExprs, expr)
		}
		if jsFiles {
			expr, err := compExp(".js$")
			checkErr(err)
			fExprs = append(fExprs, expr)
		}
	}

	fmt.Printf("Hunting for -> %s in %s\n\n", query, root)
	godirwalk.Walk(root, &godirwalk.Options{
		Unsorted: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				match := false
				if len(fExprs) == 0 {
					match = true
				} else {
					for _, expr := range fExprs {
						if expr.Match([]byte(de.Name())) {
							match = true
						}
					}
				}
				if !match {
					return nil
				}
				wg.Add(1)
				go readFile(&wg, path)
			}
			return nil
		},
		ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
	})

	wg.Wait()

	red := color.New(color.FgRed).SprintFunc()
	fmt.Printf("\n%s occurences found\n", red(fmt.Sprintf("%d", found)))

	os.Exit(found)
}
