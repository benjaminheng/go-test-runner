package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// regexps
var (
	testFuncRegexp = regexp.MustCompile(`func (Test.+)\(t \*testing.T\) {`)
	subtestRegexp  = regexp.MustCompile(`t.Run\("(.+)", func\(t \*testing.T\) {`)
)

func main() {
	flag.Parse()
	tests, err := listTests(flag.Args())
	if err != nil {
		log.Fatal(err)
	}
	selectedTest, err := selectTests(tests)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, selectedTest)
	cmd := []string{"go", "test", "-v"}
	cmd = append(cmd, flag.Args()...)
	cmd = append(cmd, "-run", selectedTest)
	cmdStr := strings.Join(cmd, " ")
	err = runShellCommand(cmdStr, nil, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func selectTests(tests []string) (selectedTest string, err error) {
	r := strings.NewReader(strings.Join(tests, "\r\n"))
	b := &bytes.Buffer{}
	err = runShellCommand("fzf", r, b)
	if err != nil {
		return "", err
	}
	selectedTest = strings.TrimSpace(b.String())
	return selectedTest, nil
}

func listTests(args []string) (tests []string, err error) {
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
				testsInFile, err := listTestsInFile(filePath)
				if err != nil {
					return err
				}
				tests = append(tests, testsInFile...)
				return nil
			})
		if err != nil {
			return nil, err
		}
	} else {
		// treat each element as a file
		for _, filePath := range args {
			base := path.Base(filePath)
			if !strings.HasSuffix(base, "_test.go") {
				continue
			}
			testsInFile, err := listTestsInFile(filePath)
			if err != nil {
				return nil, err
			}
			tests = append(tests, testsInFile...)
		}
	}
	return tests, nil
}

func listTestsInFile(filePath string) (tests []string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
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
		tests = append(tests, currentTest.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tests, nil
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

type testID struct {
	testFunc string
	subtests []string
}

func (t *testID) String() string {
	id := t.testFunc + "$"
	for _, v := range t.subtests {
		id = id + "/" + v + "$"
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
