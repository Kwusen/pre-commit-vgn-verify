package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var gitmodRegex = regexp.MustCompile(`path\s*=\s*(.*)`)
var vgnVersionRegex = regexp.MustCompile(`go\.kwusen\.ca\/vgn\s+(v.*)`)
var vgnReplaceRegex = regexp.MustCompile(`^\s*(replace)\s+go\.kwusen\.ca\/vgn.*`)

func findMatches(f *os.File, regexes ...*regexp.Regexp) [][]string {
	var found [][]string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ln := scanner.Text()
		for _, r := range regexes {
			matches := r.FindStringSubmatch(ln)

			if len(matches) >= 2 {
				found = append(found, matches)
			}
		}
	}
	return found
}

func fail(failed *bool, template string, arg ...interface{}) {
	*failed = true
	fmt.Print("- ")
	fmt.Printf(template, arg...)
	fmt.Println()
}

func main() {

	failed := false

	// Load the go module to parse the version of vgn and ensure no replace.
	modFile, err := os.Open("go.mod")
	if err != nil {
		fail(&failed, "Could not open go.mod: %s", err)
	}

	targetVersion := ""

	vgnMatches := findMatches(modFile, vgnVersionRegex, vgnReplaceRegex)
	for _, m := range vgnMatches {
		if m[1] == "replace" {
			fail(&failed, `Found "replace" in go.mod for go.kwusen.ca/vgn.

  Comment out this line:
  "%s"
  And update require block to point entry for "go.kwusen.ca/vgn" to a new 
  version, after committing and releasing, if necessary
`, m[0])
		}
		if m[0][:16] == "go.kwusen.ca/vgn" {
			targetVersion = strings.TrimSpace(m[1])
		}
	}

	if targetVersion == "" {
		fail(&failed, "Could not find version of vgn to target in go.mod.")
	}

	// Load the git modules and find the versions in each subdir.
	gmFile, err := os.Open(".gitmodules")
	if err != nil {
		fail(&failed, "Could not open .gitmodules: %s\n", err)
	}

	smPaths := findMatches(gmFile, gitmodRegex)

	for _, m := range smPaths {
		path := filepath.Join(m[1], "vgn-version.txt")
		bts, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				fail(&failed, `Submodule file not found for vgn version validation:
  "%s"

  Create and commit the file at the submodule path and copy the version 
  that is pointed at in go.mod`, path)
			} else {
				fail(&failed, "Could not open go.mod: %s\n", err)
			}
			continue
		}
		smVer := strings.TrimSpace(string(bts))
		if targetVersion != "" && smVer != targetVersion {
			fail(&failed, `Version of vgn in go.mod (%s) is not equal to "%s":
  File with incorrect version: "%s"`, targetVersion, smVer, path)
			continue
		}

		// All is good with the submodule. Make sure there are no pending changes.
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir, err = filepath.Abs(m[1])
		if err != nil {
			fail(&failed, `Failed to get absolute path for submodule: %s`, err)
		}
		files, err := cmd.CombinedOutput()
		if err != nil {
			fail(&failed, `Failed to run git status on "%s": %s`, m[1], err)
			continue
		}
		if len(files) > 0 {
			fail(&failed, "Found pending changes in \"%s\":\n\n%s", m[1], files)
		}
	}

	if failed {
		os.Exit(1)
	}
}
