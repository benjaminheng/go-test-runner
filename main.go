package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	Verbose bool
}

var config Config

// regexps
var (
	testFuncRegexp = regexp.MustCompile(`func (Test.+)\(t \*testing.T\) {`)
	subtestRegexp  = regexp.MustCompile(`t.Run\("(.+)", func\(t \*testing.T\) {`)
)

func init() {
	flag.BoolVar(&config.Verbose, "v", false, "Verbose mode")
}

func main() {
	flag.Parse()
	listTests(flag.Args())
}

func listTests(args []string) error {
	if len(args) == 1 && args[0] == "./..." {
		err := filepath.Walk("./",
			func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				base := path.Base(filePath)
				if info.IsDir() && base == "testdata" {
					// `go test` ignores directories named
					// "testdata", here we respect that.
					return nil
				}
				if !strings.HasSuffix(base, "_test.go") {
					return nil
				}
				err = listTestsInFile(filePath)
				if err != nil {
					return err
				}
				return nil
			})
		if err != nil {
			return err
		}
	} else {
		// treat each element as a file
		for _, filePath := range args {
			base := path.Base(filePath)
			if !strings.HasSuffix(base, "_test.go") {
				continue
			}
			err := listTestsInFile(filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type testID struct {
	testFunc string
	subtests []string
}

func (t *testID) String() string {
	id := t.testFunc
	for _, v := range t.subtests {
		id = id + "/" + v
	}
	return id
}

func (t *testID) AddSubtest(name string) {
	name = strings.ReplaceAll(name, " ", "_")
	t.subtests = append(t.subtests, name)
}

func (t *testID) PopSubtest() {
	if len(t.subtests) == 0 {
		return
	}
	t.subtests = t.subtests[:len(t.subtests)-1]
}

func (t *testID) Reset() {
	t.testFunc = ""
	t.subtests = nil
}

func listTestsInFile(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	currentTest := &testID{}
	var currentSubtestLevel int64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := testFuncRegexp.FindStringSubmatch(line); len(matches) > 1 {
			// in new TestXXX function
			currentTest.Reset()
			currentTest.testFunc = matches[1]
			currentSubtestLevel = 0
		} else if matches := subtestRegexp.FindStringSubmatch(line); len(matches) > 1 {
			// in subtest
			subtestLevel := countIndents(line)
			if subtestLevel == currentSubtestLevel {
				currentTest.PopSubtest()
			} else if subtestLevel < currentSubtestLevel {
				currentTest.PopSubtest()
				currentTest.PopSubtest()
			}
			currentTest.AddSubtest(matches[1])
			currentSubtestLevel = subtestLevel
		} else {
			continue
		}
		fmt.Println(currentTest)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func countIndents(text string) (count int64) {
	for _, v := range text {
		if v == '\t' {
			count++
		} else {
			return count
		}
	}
	return 0
}
