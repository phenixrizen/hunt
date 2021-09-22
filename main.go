package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/spf13/pflag"
)

var root, query, filenameRegEx, ignoreRegEx string
var cFiles, goFiles, rustFiles, rubyFiles, jsFiles, gitCommit, removeCode bool
var ignore *regexp.Regexp
var qExpr *regexp.Regexp
var fExprs = make([]*regexp.Regexp, 0)
var found = 0
var wg sync.WaitGroup

func init() {
	pflag.StringVarP(&query, "query", "q", "", "regexp to match source content")
	pflag.StringVarP(&root, "root", "r", ".", "root to start your hunt")
	pflag.StringVarP(&filenameRegEx, "name-regexp", "n", "", "regexp to match the filename")
	pflag.StringVarP(&ignoreRegEx, "ignore-regexp", "i", "^\\.git|^vendor", "regexp to ignore matching the filename")
	pflag.BoolVarP(&cFiles, "c-files", "c", false, "search for c/c++ files")
	pflag.BoolVarP(&goFiles, "go-files", "g", false, "search for Go files")
	pflag.BoolVarP(&rustFiles, "rust-files", "s", false, "search for rust files")
	pflag.BoolVarP(&rubyFiles, "ruby-files", "b", false, "search for ruby files")
	pflag.BoolVarP(&jsFiles, "js-files", "j", false, "search for JavaScript files")
	pflag.BoolVarP(&gitCommit, "git-commit", "h", false, "git the git commit details for the found line")
	pflag.BoolVarP(&removeCode, "remove-code", "v", false, "do not show the matching line of code")
	pflag.Parse()

	if pflag.NFlag() == 0 || root == "" {
		fmt.Println("hunt - a simple way to hunt for content in source files\n")
		fmt.Println("usage: hunt -query \"foo bar\" -root .\n")
		fmt.Println("flags:")
		pflag.PrintDefaults()
		os.Exit(-1)
	}
}

func readFile(wg *sync.WaitGroup, path string) {
	defer wg.Done()

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		code := strings.TrimSpace(scanner.Text())
		matches := qExpr.FindAllStringSubmatchIndex(code, -1)
		if len(matches) > 0 {
			found++
			blue := color.New(color.FgBlue).SprintFunc()
			white := color.New(color.FgWhite).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			yellow := color.New(color.FgYellow).SprintfFunc()
			output := fmt.Sprintf("%s:%s    ", blue(path), green(fmt.Sprintf("%d", i)))
			idx := 0

			if !removeCode {
				for _, match := range matches {
					output = fmt.Sprintf("%s%s%s", output, white(code[idx:match[0]]), red(code[match[0]:match[1]]))
					idx = match[1]
				}

				output = fmt.Sprintf("%s%s", output, code[idx:])
			}

			if gitCommit {
				gitLog, err := exec.Command("git", "log", fmt.Sprintf("-L%d,+1:%s", i, path), "--pretty=format:\"%h - %an, %ar : %s\"").CombinedOutput()
				if err != nil {
					log.Fatal(err)
				}
				logLines := strings.Split(string(gitLog), "\n")
				if len(logLines) == 0 {
					log.Fatal("error getting git log")
				}
				output = fmt.Sprintf("%s %s %s", output, green("->"), yellow("%s", logLines[0]))
			}

			fmt.Printf("%s\n", output)
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

	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	ignoreExpr, err := compExp(ignoreRegEx)
	checkErr(err)

	fmt.Printf("Hunting for -> '%s' in '%s'\n\n", red(query), blue(root))
	godirwalk.Walk(root, &godirwalk.Options{
		Unsorted: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.IsDir() && !ignoreExpr.MatchString(path) {
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

	fmt.Printf("\n%s occurrences found\n", red(fmt.Sprintf("%d", found)))

	os.Exit(found)
}
