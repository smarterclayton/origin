package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	errorsutil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

func main() {
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	testBinaries, err := getTestBinaries(streams)
	if err != nil {
		klog.Fatal(err)
	}
	if verbose() {
		for _, testBinary := range testBinaries {
			klog.Infof("Found test binary: %s", testBinary)
		}
	}

	for _, testBinary := range testBinaries {
		testCommand := exec.Command(testBinary, os.Args[1:]...)
		testCommand.Stdout = streams.Out
		testCommand.Stderr = streams.ErrOut
		testCommand.Stdin = streams.In

		if err := testCommand.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}
	}
}

// verbose returns true if additional output from this binary should be printed.
func verbose() bool {
	return len(os.Getenv("OPENSHIFT_TESTS_DEBUG_PLUGINS")) > 0
}

func getTestBinaries(streams genericclioptions.IOStreams) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	executablePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	executableDir := filepath.Dir(executablePath)
	testBinaryDirectories := []string{}
	testBinaryDirectories = append(testBinaryDirectories, cwd)
	testBinaryDirectories = append(testBinaryDirectories, executableDir)
	testBinaryDirectories = append(testBinaryDirectories, filepath.SplitList(os.Getenv("PATH"))...)

	var warnings []string
	var errors []error
	testBinaries := sets.NewString()

	verifier := &CommandOverrideVerifier{
		seenPlugins: make(map[string]string),
	}
	for _, dir := range uniquePathsList(testBinaryDirectories) {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			errors = append(errors, fmt.Errorf("error: unable to read directory %q in your PATH: %v", dir, err))
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !nameMatchesTest(f.Name()) {
				continue
			}

			testBinaryPath := filepath.Join(dir, f.Name())
			isSymlink, err := evalSymlink(testBinaryPath)
			if err != nil {
				klog.Errorf("unable to evaluate symlink: %v", err)
			}

			if testBinaries.Has(testBinaryPath) || isSymlink {
				continue
			}
			testBinaries.Insert(testBinaryPath)

			if errs := verifier.Verify(streams, filepath.Join(dir, f.Name())); len(errs) != 0 {
				for _, err := range errs {
					warnings = append(warnings, fmt.Sprintf("%s: %v", err))
				}
			}
		}
	}
	if len(warnings) > 0 {
		if len(warnings) == 1 {
			errors = append(errors, fmt.Errorf("error: plugin warning: %s", warnings[0]))
		} else {
			errors = append(errors, fmt.Errorf("error: plugin warnings:\n%s", strings.Join(warnings, "\n")))
		}
	}
	if len(testBinaries) == 0 {
		errors = append(errors, fmt.Errorf("error: unable to find any openshift-tests plugins in your PATH"))
	}
	if len(errors) > 0 {
		return nil, errorsutil.NewAggregate(errors)
	}

	return testBinaries.List(), nil
}

type CommandOverrideVerifier struct {
	//root        *cobra.Command
	seenPlugins map[string]string
}

// evalSymlink returns true if provided path is a symlink
func evalSymlink(path string) (bool, error) {
	link, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false, err
	}
	if len(link) != 0 {
		if link != path {
			return true, nil
		}
	}
	return false, nil
}

// Verify implements PathVerifier and determines if a given path
// is valid depending on whether or not it overwrites an existing
// kubectl command path, or a previously seen plugin.
func (v *CommandOverrideVerifier) Verify(streams genericclioptions.IOStreams, path string) []error {
	// extract the plugin binary name
	segs := strings.Split(path, "/")
	binName := segs[len(segs)-1]

	cmdPath := strings.Split(binName, "-")
	if len(cmdPath) > 1 {
		// the first argument is always "kubectl" for a plugin binary
		cmdPath = cmdPath[1:]
	}

	errors := []error{}

	if isExec, err := isExecutable(path); err == nil && !isExec {
		errors = append(errors, fmt.Errorf("error: %s identified as a kubectl plugin, but it is not executable", path))
	} else if err != nil {
		errors = append(errors, fmt.Errorf("error: unable to identify %s as an executable file: %v", path, err))
	}

	if existingPath, ok := v.seenPlugins[binName]; ok {
		if verbose() {
			klog.Warningf("%s is overshadowed by a similarly named plugin: %s\n", path, existingPath)
		}
	} else {
		v.seenPlugins[binName] = path
	}

	return errors
}

func isExecutable(fullPath string) (bool, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}

	if m := info.Mode(); !m.IsDir() && m&0111 != 0 {
		return true, nil
	}

	return false, nil
}

// uniquePathsList deduplicates a given slice of strings without
// sorting or otherwise altering its order in any way.
func uniquePathsList(paths []string) []string {
	seen := map[string]bool{}
	newPaths := []string{}
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		newPaths = append(newPaths, p)
	}
	return newPaths
}

func nameMatchesTest(filepath string) bool {
	for _, prefix := range []string{"openshift-tests"} {
		if !strings.HasPrefix(filepath, prefix+"-") {
			continue
		}
		return true
	}
	return false
}
