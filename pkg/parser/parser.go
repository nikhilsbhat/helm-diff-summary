package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var headerRegex = regexp.MustCompile(
	`^([a-zA-Z0-9-]+),\s*([a-zA-Z0-9-]+),\s*([a-zA-Z0-9-]+).*?(has been added|has changed|has been removed):?$`,
)

// Parse parses the helm diff output to get the summary.
func Parse(r io.Reader) ([]ResourceDiff, error) {
	scanner := bufio.NewScanner(r)

	var (
		resources []ResourceDiff
		current   *ResourceDiff
	)

	finalize := func() {
		if current == nil {
			return
		}

		// Terraform-style logical change counting
		switch current.ChangeType {
		case Create:
			current.ChangedLines = current.Additions

		case Delete:
			current.ChangedLines = current.Deletions

		case Update:
			current.ChangedLines = current.Additions
		}

		current.Category = detectCategory(current.Kind)
		current.Severity = detectSeverity(current)

		resources = append(resources, *current)
	}

	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		// Example:
		//
		// default, sample, Deployment (apps) has been added:
		//
		matches := headerRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			finalize()

			namespace := matches[1]
			name := matches[2]
			kind := matches[3]
			actionText := matches[4]

			var action ChangeType

			switch {
			case strings.Contains(actionText, "added"):
				action = Create

			case strings.Contains(actionText, "removed"):
				action = Delete

			default:
				action = Update
			}

			current = &ResourceDiff{
				Namespace:  namespace,
				Name:       name,
				Kind:       kind,
				ChangeType: action,
			}

			continue
		}

		if current == nil {
			continue
		}

		// Count meaningful additions
		if strings.HasPrefix(raw, "+") &&
			!strings.HasPrefix(raw, "+++") {
			content := strings.TrimSpace(
				strings.TrimPrefix(raw, "+"),
			)

			if content != "" &&
				!strings.HasPrefix(content, "#") &&
				!strings.HasPrefix(content, "@@") {
				current.Additions++
			}
		}

		// Count meaningful deletions
		if strings.HasPrefix(raw, "-") &&
			!strings.HasPrefix(raw, "---") {
			content := strings.TrimSpace(
				strings.TrimPrefix(raw, "-"),
			)

			if content != "" &&
				!strings.HasPrefix(content, "#") &&
				!strings.HasPrefix(content, "@@") {
				current.Deletions++
			}
		}
	}

	finalize()

	return resources, scanner.Err()
}
