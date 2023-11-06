package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type Matrix struct {
	Include []Changeset `json:"include"`
}

func main() {
	// Define the regex pattern
	pattern := `(\w+-){2}\w+\.md`
	regex := regexp.MustCompile(pattern)

	// Specify the directory path
	directory := ".changeset/"

	// Get a list of files in the directory
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println("Couldn't open directory .changeset")
		return
	}

	parsedChangesets := make([]Changeset, 0)

	// Loop through the files
	for _, file := range files {
		// Check if the file matches the regex pattern
		if regex.MatchString(file.Name()) {
			fmt.Sprintln("Processing changeset", file.Name())
			parsedChangesets = append(parsedChangesets, parseChangesetFile(filepath.Join(directory, file.Name())))
		}
	}

	matrix := Matrix{Include: parsedChangesets}

	jsonOutput, _ := json.Marshal(matrix)

	jsonOutputStr := string(jsonOutput)
	jsonOutputStr = "release-note-matrix=" + jsonOutputStr

	// magic env from github actions
	fileName := os.Getenv("GITHUB_OUTPUT")
	_ = os.WriteFile(fileName, []byte(jsonOutputStr), 0441)
}

type Version struct {
	Pkg     string `json:"package"`
	Upgrade string `json:"upgrade"`
}

type Changeset struct {
	Message  string    `json:"message"`
	Versions []Version `json:"versions"`
}

func parseChangesetFile(filePath string) Changeset {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return Changeset{}
	}

	fileContentStr := string(fileContent)

	versionPattern := regexp.MustCompile("'(.+)': (patch|minor|major)")
	versions := versionPattern.FindAllStringSubmatch(fileContentStr, -1)

	changeset := Changeset{}

	for _, version := range versions {
		changeset.Versions = append(changeset.Versions, Version{
			Pkg:     version[1],
			Upgrade: version[2],
		})
	}

	messagePattern := regexp.MustCompile("(?m)^---[\\s\\S]+---\\n\\n([\\s\\S]+)$")
	message := messagePattern.FindStringSubmatch(fileContentStr)[1]
	changeset.Message = message

	return changeset
}
