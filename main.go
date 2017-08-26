package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/github"
)

// Config holds TOML file data
type Config struct {
	Repositories []string `toml:"repositories"`
	Labels       []Label  `toml:"label"`
}

// Label represents an Issue label
type Label struct {
	Name     string   `toml:"name"`
	Color    string   `toml:"color"`
	Mappings []string `toml:"mappings,omitempty"`
	Delete   bool     `toml:"delete,omitempty"`
}

func main() {

	// Load config file
	tomlData, err := ioutil.ReadFile("config.toml")
	if err != nil {
		fmt.Println("Unable to load config file.")
		os.Exit(1)
	}

	var conf Config
	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		fmt.Println("Error decoding toml.")
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_AUTH_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	for _, repo := range conf.Repositories {
		if conf.Labels != nil {
			for _, label := range conf.Labels {
				fmt.Printf("Working on label '%v' for %s.\n", label.Name, repo)
				if err := ensureLabel(ctx, repo, label, client); err != nil {
					fmt.Printf("Error, label '%v' on repository %s.\n    %v\n", label.Name, repo, err)
				}
			}
		}
	}
}

// Some ground rules:
// We'll never rename a label
// To do mappings, we'll create the new label, and then for each issue
// with the old label, we'll add the new label, then remove the old one.
func ensureLabel(ctx context.Context, repo string, label Label, gh *github.Client) error {

	splitRepo := strings.Split(repo, "/")
	owner := splitRepo[0]
	repo = splitRepo[1]
	mappings := label.Mappings

	opts := &github.ListOptions{}
	allLabels, _, err := gh.Issues.ListLabels(ctx, owner, repo, opts)
	if err != nil {
		return err
	}

	// Delete labels with the delete flag
	if label.Delete && isLabelInSlice(label.Name, allLabels) {
		resp, err := gh.Issues.DeleteLabel(ctx, owner, repo, label.Name)
		if err != nil {
			return github.CheckResponse(resp.Response)
		}

		fmt.Printf("Deleted label: %s\n", label.Name)
		return nil
	} else if label.Delete && !isLabelInSlice(label.Name, allLabels) {
		return nil
	}

	// Create labels if they don't exist; otherwise make color match
	if !isLabelInSlice(label.Name, allLabels) {
		_, _, err := gh.Issues.CreateLabel(ctx, owner, repo,
			&github.Label{Name: &label.Name, Color: &label.Color})
		if err != nil {
			return err
		}

		fmt.Println("Successfully added label.")
	} else {
		resp, _, err := gh.Issues.GetLabel(ctx, owner, repo, label.Name)
		if err != nil {
			return err
		}

		if resp.Color != &label.Color {
			_, _, err := gh.Issues.EditLabel(ctx, owner, repo, label.Name,
				&github.Label{Name: &label.Name, Color: &label.Color})
			if err != nil {
				return err
			}

		}
	}

	// Map new labels onto issues with old labels
	for _, oldLabel := range mappings {
		if isLabelInSlice(oldLabel, allLabels) {
			fmt.Printf("Working on %s mapping for %s.\n", oldLabel, label.Name)
			count := 0
			issues, err := issuesWith(ctx, owner, repo, "label", oldLabel, gh)
			if err != nil {
				return err
			}

			if len(issues) > 0 {
				for _, k := range issues {

					fmt.Printf("Updating issue %d.\n", k.Number)
					fmt.Printf("Adding label: %s.\n", label.Name)
					_, _, err := gh.Issues.AddLabelsToIssue(ctx, owner, repo, k.GetNumber(), []string{label.Name})
					if err != nil {
						return err
					}

					fmt.Printf("Removing label %s\n.", oldLabel)
					_, err = gh.Issues.RemoveLabelForIssue(ctx, owner, repo, k.GetNumber(), oldLabel)
					if err != nil {
						return err
					}

					count++
					fmt.Printf("-----\n")

				}
			}

			// Check once more before deleting label
			remaining, err := issuesWith(ctx, owner, repo, "label", oldLabel, gh)
			if err != nil {
				return err
			}

			if len(remaining) == 0 || len(remaining) == count {
				fmt.Printf("Deleting label %s from %s.\n", oldLabel, repo)
				_, err := gh.Issues.DeleteLabel(ctx, owner, repo, oldLabel)
				if err != nil {
					return err
				}

			} else {
				fmt.Printf("There are remaining issues returned from search - you should manually check that the %s label has no tickets assigned (or rerun this script in a few minutes time)\n", oldLabel)
			}
		}

	}

	return nil
}

func issuesWith(ctx context.Context, owner, repo, kind, label string, gh *github.Client) ([]github.Issue, error) {
	query := fmt.Sprintf("repo:%s/%s %s:\"%s\"", owner, repo, kind, label)

	resp, _, err := gh.Search.Issues(ctx, query, &github.SearchOptions{})
	if err != nil {
		fmt.Println("help")
	}

	return resp.Issues, nil
}

func isLabelInSlice(a string, list []*github.Label) bool {
	for _, b := range list {
		if b.GetName() == a {
			return true
		}
	}
	return false
}
