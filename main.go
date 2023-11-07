package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/slack-go/slack"
)

type Matrix struct {
	Include []Changeset `json:"include"`
}

func main() {
	// This is passed in by action inputs
	token := os.Args[1]
	channelId := os.Args[2]
	headline := os.Args[3]

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

	api := slack.New(token, slack.OptionDebug(true))

	releaseNoteBlocks := []slack.Block{}

	releaseNoteBlocks = append(releaseNoteBlocks,
		slack.HeaderBlock{
			Type: "header",
			Text: &slack.TextBlockObject{
				Type:  "plain_text",
				Text:  headline,
				Emoji: true,
			},
		},
	)

	for _, releaseNote := range parsedChangesets {
		releaseNoteBlocks = append(releaseNoteBlocks, slack.ContextBlock{
			Type: "context",
			ContextElements: slack.ContextElements{
				Elements: []slack.MixedElement{
					slack.TextBlockObject{
						Type: "mrkdwn",
						Text: releaseNote.Message,
					},
				},
			},
		})
	}

	_, _, _, err = api.SendMessage(channelId, slack.MsgOptionBlocks(
		releaseNoteBlocks...,
	))

	if err != nil {
		fmt.Println(err)
		return
	}
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
