package policy

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
)

func storagePolicies() []Config {
	// ------------------------------------------------------------
	// Storage resources
	// ------------------------------------------------------------
	return []Config{
		{
			Name:     "storage-update",
			Category: parser.Storage,
			Action:   parser.Update,
			Severity: parser.High,
			Message:  "storage resource updated",
		},

		{
			Name:     "storage-deletion",
			Category: parser.Storage,
			Action:   parser.Delete,
			Severity: parser.Critical,
			Message:  "storage resource deleted",
		},
	}
}

func platformPolicies() []Config {
	// ------------------------------------------------------------
	// Platform resource modifications
	// ------------------------------------------------------------
	return []Config{
		{
			Name:     "platform-update",
			Category: parser.Platform,
			Action:   parser.Update,
			Severity: parser.High,
			Message:  "platform resource updated",
		},

		{
			Name:     "platform-deletion",
			Category: parser.Platform,
			Action:   parser.Delete,
			Severity: parser.Critical,
			Message:  "platform resource deleted",
		},
	}
}

func networkPolicies() []Config {
	// ------------------------------------------------------------
	// Networking changes
	// ------------------------------------------------------------
	return []Config{
		{
			Name:     "networking-update",
			Category: parser.Networking,
			Action:   parser.Update,
			Severity: parser.High,
			Message:  "networking resource updated",
		},

		{
			Name:     "networking-create",
			Category: parser.Networking,
			Action:   parser.Create,
			Severity: parser.Medium,
			Message:  "new networking resource created",
		},
	}
}

func sensitiveNamespacePolicies() []Config {
	namespaces := []string{
		"production",
		"prod",
		"kube-system",
		"istio-system",
		"linkerd",
		"linkerd-viz",
		"cert-manager",
		"crossplane-system",
		"monitoring",
		"observability",
		"karpenter",
		"argocd",
	}

	policies := make([]Config, 0, len(namespaces))

	for _, namespace := range namespaces {
		policies = append(
			policies,
			Config{
				Name: fmt.Sprintf(
					"sensitive-namespace-%s",
					namespace,
				),
				Namespace: namespace,
				Severity:  parser.High,
				Message:   "change detected in sensitive namespace",
			},
		)
	}

	return policies
}

func criticalPlatformPolicies() []Config {
	kinds := []string{
		"CustomResourceDefinition",
		"Namespace",
		"MutatingWebhookConfiguration",
		"ValidatingWebhookConfiguration",
		"ClusterRole",
		"ClusterRoleBinding",
	}

	policies := make([]Config, 0, len(kinds))

	for _, kind := range kinds {
		policies = append(
			policies,
			Config{
				Name: fmt.Sprintf(
					"critical-platform-resource-%s",
					strings.ToLower(kind),
				),
				Kind:     kind,
				Severity: parser.Critical,
				Message:  "critical platform resource modified",
			},
		)
	}

	return policies
}
