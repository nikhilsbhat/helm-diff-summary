package parser

type PolicyViolation struct {
	Name     string   `yaml:"name,omitempty"     json:"name,omitempty"`
	Severity Severity `yaml:"severity,omitempty" json:"severity,omitempty"`
	Resource string   `yaml:"resource,omitempty" json:"resource,omitempty"`
	Message  string   `yaml:"message,omitempty"  json:"message,omitempty"`
}

func EvaluatePolicies(resources []ResourceDiff) []PolicyViolation {
	violations := make([]PolicyViolation, 0, len(resources))

	for _, resource := range resources {
		violations = append(
			violations,
			evaluateResourcePolicies(resource)...,
		)
	}

	return violations
}

func HasViolationsAtOrAbove(violations []PolicyViolation, severity Severity) bool {
	for _, violation := range violations {
		if severityRank(violation.Severity) >= severityRank(severity) {
			return true
		}
	}

	return false
}

func evaluateResourcePolicies(resource ResourceDiff) []PolicyViolation {
	violations := make([]PolicyViolation, 0)

	// ------------------------------------------------------------
	// Block deletions
	// ------------------------------------------------------------

	if resource.ChangeType == Delete {
		violations = append(
			violations,
			PolicyViolation{
				Name:     "resource-deletion",
				Severity: Critical,
				Resource: resource.Name,
				Message:  "resource deletion detected",
			},
		)
	}

	// ------------------------------------------------------------
	// Block platform resource changes
	// ------------------------------------------------------------

	if resource.Category == Platform &&
		resource.ChangeType == Delete {
		violations = append(
			violations,
			PolicyViolation{
				Name:     "platform-deletion",
				Severity: Critical,
				Resource: resource.Name,
				Message:  "platform resource deletion detected",
			},
		)
	}

	// ------------------------------------------------------------
	// High-risk networking changes
	// ------------------------------------------------------------

	if resource.Category == Networking &&
		resource.ChangeType == Update {
		violations = append(
			violations,
			PolicyViolation{
				Name:     "networking-update",
				Severity: High,
				Resource: resource.Name,
				Message:  "networking resource updated",
			},
		)
	}

	// ------------------------------------------------------------
	// Production namespace escalation
	// ------------------------------------------------------------

	switch resource.Namespace {
	case "production",
		"prod",
		"kube-system",
		"istio-system":
		violations = append(
			violations,
			PolicyViolation{
				Name:     "sensitive-namespace",
				Severity: High,
				Resource: resource.Name,
				Message:  "change detected in sensitive namespace",
			},
		)
	}

	return violations
}

func severityRank(severity Severity) int {
	const (
		defaultSeverityRank0 = 0
		defaultSeverityRank1 = 1
		defaultSeverityRank2 = 2
		defaultSeverityRank3 = 3
		defaultSeverityRank4 = 4
	)

	switch severity {
	case Low:
		return defaultSeverityRank1
	case Medium:
		return defaultSeverityRank2
	case High:
		return defaultSeverityRank3
	case Critical:
		return defaultSeverityRank4
	}

	return defaultSeverityRank0
}
