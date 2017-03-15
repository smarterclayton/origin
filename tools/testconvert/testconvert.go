package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/openshift/imagebuilder"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

type test struct {
	before, after bool
	name          string
	lines         []string
	children      []*test
}

func main() {
	var from string
	flag.StringVar(&from, "from", "", "file to convert to a test file")
	flag.Parse()
	if len(from) == 0 {
		glog.Fatal("You must specify -from to the path to a script")
	}
	name, dir := filepath.Base(from), filepath.Dir(from)
	if !strings.HasSuffix(name, ".sh") {
		glog.Fatal("Only supports files with a .sh extension")
	}
	name = strings.TrimSuffix(name, ".sh")
	dirs := strings.Split(dir, string(filepath.Separator))

	out, err := ioutil.ReadFile(from)
	if err != nil {
		glog.Fatal(err)
	}

	if len(dirs) == 0 {
		dirs = []string{"test"}
	}

	lines := strings.Split(string(out), "\n")
	tests, remaining := parse(lines)
	if last := trimLines(remaining); len(last) > 0 {
		tests = append(tests, &test{after: true, name: "last", lines: last})
	}

	fmt.Printf(`
package %s

import (
	"testing"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("[Serial][Feature:Command] %s %s", func() {
	defer g.GinkgoRecover()
`, dirs[len(dirs)-1], strings.Join(dirs, " "), name)

	describe(tests)

	fmt.Printf(`
})
`)
}

func describe(tests []*test) {
	for _, test := range tests {
		if len(test.lines) > 0 {
			if test.before {
				fmt.Printf(`
	g.BeforeEach(func() {
%s
	})
`, block(each(test.lines, bash, indent("\t  "))))
				continue
			}

			fmt.Printf(`
	g.Describe("%s", func() {
%s
	})
`, test.name, block(each(test.lines, bash, indent("\t  "))))
		}
		if len(test.children) > 0 {
			describe(test.children)
		}
	}
}

func nextDirective(lines []string) (ok, start bool, name string, position int) {
	const begin = "os::test::junit::declare_suite_start "
	const end = "os::test::junit::declare_suite_end"
	for i, s := range lines {
		switch {
		case strings.HasPrefix(s, begin):
			testName, err := strconv.Unquote(strings.TrimPrefix(s, begin))
			if err != nil {
				glog.Fatalf("Can't unquote name from line %d: %s", i+1, s)
			}
			return true, true, testName, i
		case strings.HasPrefix(s, end):
			return true, false, "", i
		}
	}
	return false, false, "", 0
}

func parse(lines []string) ([]*test, []string) {
	fmt.Printf("// LINES ENTER %d\n", len(lines))
	var children []*test
	var last *test
	remaining := lines
	for {
		fmt.Printf("// SCAN %v\n", remaining)
		ok, start, testName, at := nextDirective(remaining)
		switch {

		case !ok && last != nil:
			fmt.Printf("// NO MATCH, CLOSING %s\n", last.name)
			last.lines = trimLines(remaining)
			return children, nil

		case !ok:
			fmt.Printf("// NO MATCH\n")
			return children, remaining

		case start && last != nil:
			fmt.Printf("// NESTED ENTER (%s): %d: %v\n", last.name, at, remaining[:at])
			// add to the active child, fill out the body and recurse
			last.lines = trimLines(remaining[:at])
			last.children, remaining = parse(remaining[at:])
			fmt.Printf("// NESTED EXIT  (%s): %v\n", last.name, remaining)

		case start:
			if at > 0 {
				fmt.Printf("// TEST PREAMBLE %d\n", at)
				children = append(children, &test{before: true, name: fmt.Sprintf("section %d", len(children)+1), lines: trimLines(remaining[:at])})
			}
			// start filling out a new test
			last = &test{name: testName}
			children = append(children, last)
			fmt.Printf("// TEST STARTED %s\n", testName)
			remaining = remaining[at+1:]

		case last != nil:
			// test has completed
			fmt.Printf("// TEST ENDED %s (%d)\n", last.name, at)
			if last.lines == nil {
				last.lines = trimLines(remaining[:at])
			} else {
				fmt.Printf("// TEST STUB %d\n", at)
				children = append(children, &test{after: true, name: fmt.Sprintf("section %d", len(children)+1), lines: trimLines(remaining[:at])})
			}
			last = nil

			remaining = remaining[at+1:]

		default:
			fmt.Printf("// LINES EXIT %d (too many end): %v\n", len(children), remaining)
			return children, remaining[at:]
		}
	}
}

func splitOnLine(lines []string, prefix string) (before, contains []string) {
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], prefix) {
			return lines[:i], lines[i:]
		}
	}
	return lines, nil
}

func trimLines(lines []string) []string {
	var result []string
	for i, s := range lines {
		if len(strings.TrimSpace(s)) != 0 {
			result = lines[i:]
			break
		}
	}
	for i := len(result) - 1; i >= 0; i-- {
		if len(strings.TrimSpace(result[i])) != 0 {
			result = result[:i+1]
			break
		}
	}
	return result
}

func transformBody(lines []string) []string {
	return lines
}

func bash(line string) string {
	switch {
	case strings.HasPrefix(line, "# "):
		return "// " + strings.TrimPrefix(line, "# ")
	case strings.HasPrefix(line, "os::cmd::expect_success "):
		arg := strings.TrimPrefix(line, "os::cmd::expect_success ")
		args, err := imagebuilder.ProcessWords(arg, nil)
		if err != nil {
			panic(err)
		}
		if len(args) != 1 {
			panic(args)
		}
		return fmt.Sprintf("o.Expect(bash.Exec(%q)).NotTo(o.HaveOccurred())", args[0])
	case strings.HasPrefix(line, "os::cmd::expect_failure_and_text "):
		arg := strings.TrimPrefix(line, "os::cmd::expect_failure_and_text ")
		args, err := imagebuilder.ProcessWords(arg, nil)
		if err != nil {
			panic(err)
		}
		if len(args) != 2 {
			panic(args)
		}
		return fmt.Sprintf("o.Expect(bash.ExecFailWithText(%q, %q)).NotTo(o.HaveOccurred())", args[0], args[1])
	case strings.HasPrefix(line, "os::cmd::expect_success_and_text "):
		arg := strings.TrimPrefix(line, "os::cmd::expect_success_and_text ")
		args, err := imagebuilder.ProcessWords(arg, nil)
		if err != nil {
			panic(err)
		}
		if len(args) != 2 {
			return fmt.Sprintf("// ERROR: expected 2 arguments from: \n// %s", line)
		}
		return fmt.Sprintf("o.Expect(bash.ExecWithText(%q, %q)).NotTo(o.HaveOccurred())", args[0], args[1])
	case strings.HasPrefix(line, "os::cmd::expect_success_and_not_text "):
		arg := strings.TrimPrefix(line, "os::cmd::expect_success_and_not_text ")
		args, err := imagebuilder.ProcessWords(arg, nil)
		if err != nil {
			panic(err)
		}
		if len(args) != 2 {
			panic(args)
		}
		return fmt.Sprintf("o.Expect(bash.ExecWithoutText(%q, %q)).NotTo(o.HaveOccurred())", args[0], args[1])
	}
	return "// " + line
}

func indent(indent string) func(string) string {
	return func(s string) string {
		return indent + s
	}
}

func each(lines []string, fns ...func(string) string) []string {
	var result []string
	for _, s := range lines {
		for _, fn := range fns {
			s = fn(s)
		}
		result = append(result, s)
	}
	return result
}

func block(lines []string) string {
	return strings.Join(lines, "\n")
}
