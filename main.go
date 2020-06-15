package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"path/filepath"
	"strings"
	"io"
)

func main() {
	var (
		reRoot = regexp.MustCompile(`https://.+/internal-resources(.+)`)
		root = "./docs"
	)

	// Open our jsonFile
	jsonFile, err := os.Open("pages.json")
	// if we os.Open returns an error then handle it
	if err != nil { fmt.Println(err) }
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	type Page struct {
		Link, Frontmatter, Markdown string
	}
	dec := json.NewDecoder(jsonFile)

	// read open bracket
	t, err := dec.Token()
	if err != nil { fmt.Println(err); fmt.Println(t) }

	// while the array contains values
	for dec.More() {
		var p Page
		// decode an array value (Page)
		err := dec.Decode(&p)
		if err != nil { fmt.Println(err) }

		// For each object, get its details
		path := reRoot.ReplaceAllString(fmt.Sprintf("%v", p.Link), root + "$1")
		content := fmt.Sprintf("---\n%v\n---\n%v\n", p.Frontmatter, p.Markdown)

		_ = os.MkdirAll(path, 0775) // Create a directory
		file, err := os.Create(path + ".md") // Create a md file
		if err != nil { fmt.Println(err) }
		file.WriteString(content)
		file.Close()
	}

	ReOrgFiles(root) // Reorganize the files for VuePress
}

func ReOrgFiles(root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { // prevent panic by handling failure accessing a path
			// We moved some files, so ignore file access failures
			return nil
			// fmt.Printf("failure accessing %q: %v\n", path, err)
		}
		// Process only directories
		if info.IsDir() {
			dir, err := os.Open(path)
			if err != nil { fmt.Println(err) }
			_, err = dir.Readdirnames(1)
			defer dir.Close()
			if err == io.EOF { // If this is an empty dir, delete it
				os.RemoveAll(path)
			} else { // otherwise, move the md file with the same name into it
				mdFile := path + ".md"
				if FileExists(mdFile) { // If this directory has a matching .md file
					// rename the file to directory/README.md
					readmeFile := strings.TrimSuffix(mdFile,".md") + "/README.md"
					err := os.Rename(mdFile, readmeFile)
					if err != nil { fmt.Println(err) }
				}
			}
		}
		return nil
	})
	return err
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) { return false }
	return !info.IsDir()
}
