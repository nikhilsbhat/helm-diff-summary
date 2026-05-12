package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	headerRegex = regexp.MustCompile(
		`^([a-zA-Z0-9-]+),\s*([a-zA-Z0-9-]+),\s*([a-zA-Z0-9-]+).*?(has been added|has changed|has been removed):?$`,
	)
	categoryMappings = map[Category]map[string]struct{}{
		Workload: resourceSet(
			"Deployment",
			"StatefulSet",
			"DaemonSet",
			"ReplicaSet",
			"ReplicationController",
			"Job",
			"CronJob",
			"Pod",
			"HorizontalPodAutoscaler",
			"VerticalPodAutoscaler",
			"Rollout",
			"AnalysisRun",
			"Experiment",
			"AnalysisTemplate",
			"ClusterAnalysisTemplate",
			"Notebook",
			"PyTorchJob",
			"TFJob",
			"MPIJob",
			"XGBoostJob",
			"PaddleJob",
			"RayCluster",
			"RayJob",
			"RayService",
			"SparkApplication",
			"ScheduledSparkApplication",
			"TaskRun",
			"PipelineRun",
			"Workflow",
			"WorkflowTemplate",
			"CronWorkflow",
			"CloneSet",
			"AdvancedStatefulSet",
			"UnitedDeployment",
			"BroadcastJob",
		),
		Networking: resourceSet(
			"Service",
			"Ingress",
			"Gateway",
			"HTTPRoute",
			"VirtualService",
			"NetworkPolicy",
			"Endpoints",
			"EndpointSlice",
			"IngressRoute",
			"IngressRouteTCP",
			"IngressRouteUDP",
			"Middleware",
			"MiddlewareTCP",
			"TLSOption",
			"TLSStore",
			"TraefikService",
			"ServersTransport",
			"ServersTransportTCP",
			"DestinationRule",
			"AuthorizationPolicy",
			"PeerAuthentication",
			"TCPRoute",
			"GRPCRoute",
			"ControlPlane",
			"DataPlane",
			"GatewayConfiguration",
			"KongPlugin",
			"KongClusterPlugin",
			"KongConsumer",
			"KongConsumerGroup",
			"KongIngress",
			"ServiceProfile",
			"Server",
			"ServerAuthorization",
			"MeshTLSAuthentication",
			"NetworkAuthentication",
			"TLSRoute",
			"TrafficSplit",
			"Link",
			"EgressNetwork",
			"Policy",
			"Authorization",
		),
		Security: resourceSet(
			"ClusterRole",
			"ClusterRoleBinding",
			"Role",
			"RoleBinding",
			"ServiceAccount",
			"PodSecurityPolicy",
			"Certificate",
			"Issuer",
			"ClusterIssuer",
			"CertificateRequest",
			"ExternalSecret",
			"SecretStore",
			"ClusterSecretStore",
			"PushSecret",
			"VaultAuth",
			"VaultConnection",
			"VaultStaticSecret",
			"VaultDynamicSecret",
			"VaultPKISecret",
			"VaultAuthGlobal",
			"VaultPolicy",
			"SecretProviderClass",
			"SecretProviderClassPodStatus",
			"SealedSecret",
		),
		Storage: resourceSet(
			"PersistentVolume",
			"PersistentVolumeClaim",
			"StorageClass",
			"VolumeAttachment",
			"CSIDriver",
			"CSINode",
			"CSIStorageCapacity",
			"VolumeSnapshot",
			"VolumeSnapshotClass",
			"VolumeSnapshotContent",
			"Backup",
			"Restore",
			"Schedule",
			"BackupStorageLocation",
			"VolumeSnapshotLocation",
			"DeleteBackupRequest",
			"PodVolumeBackup",
			"PodVolumeRestore",
			"CephCluster",
			"CephBlockPool",
			"CephFilesystem",
			"CephObjectStore",
			"CephObjectStoreUser",
			"CephFilesystemSubVolumeGroup",
			"CephNFS",
			"Volume",
			"Engine",
			"Replica",
			"BackingImage",
			"BackupVolume",
			"RecurringJob",
			"SystemBackup",
			"CStorPoolCluster",
			"CStorVolume",
			"BlockDevice",
			"DiskPool",
			"Tenant",
			"PolicyBinding",
			"MongoDBCommunity",
			"PerconaServerMongoDB",
			"PostgresCluster",
			"PGUpgrade",
			"Elasticsearch",
		),
		Platform: resourceSet(
			"Namespace",
			"CustomResourceDefinition",
			"MutatingWebhookConfiguration",
			"ValidatingWebhookConfiguration",
			"APIService",
			"PriorityClass",
			"Provider",
			"Configuration",
			"Function",
			"Composition",
			"CompositeResourceDefinition",
			"Application",
			"AppProject",
			"ApplicationSet",
			"Prometheus",
			"Alertmanager",
			"ServiceMonitor",
			"Receiver",
			"PodMonitor",
			"Probe",
			"PrometheusRule",
			"ScrapeConfig",
			"ThanosRuler",
			"Thanos",
			"Grafana",
			"NodePool",
			"EC2NodeClass",
			"NodeClaim",
			"Provisioner",
			"AWSNodeTemplate",
			"LinkerdControlPlane",
			"LinkerdDataPlane",
			"IngressClassParams",
			"TargetGroupBinding",
			"GatewayClass",
			"ReferenceGrant",
			"AzureIngressProhibitedTarget",
			"IngressClassParameters",
			"Route",
			"ScaledObject",
			"ScaledJob",
			"TriggerAuthentication",
			"ClusterTriggerAuthentication",
		),
		Config: resourceSet(
			"ConfigMap",
			"Secret",
		),
	}

	resourceCategoryMap = buildCategoryMap()
)

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

func detectCategory(kind string) Category {
	if category, ok := resourceCategoryMap[kind]; ok {
		return category
	}

	return Unknown
}

func detectSeverity(resource *ResourceDiff) Severity {
	const (
		defaultChangedLength50  = 50
		defaultChangedLength200 = 200
		defaultScore2           = 2
		defaultScore3           = 3
		defaultScore4           = 4
		defaultScore5           = 5
		defaultScore6           = 6
		defaultScore9           = 9
	)

	score := 0
	// ------------------------------------------------------------
	// Action scoring
	// ------------------------------------------------------------

	switch resource.ChangeType {
	case Create:
		score++

	case Update:
		score += defaultScore2

	case Delete:
		score += defaultScore5
	}

	// ------------------------------------------------------------
	// Category scoring
	// ------------------------------------------------------------

	switch resource.Category {
	case Networking:
		score += defaultScore2

	case Security:
		score += defaultScore3

	case Storage:
		score += defaultScore3

	case Platform:
		score += defaultScore4

	case Workload:
		score++

	case Config:
		score++

	case Unknown:
		score += defaultScore2
	}

	// ------------------------------------------------------------
	// Namespace escalation
	// ------------------------------------------------------------

	switch resource.Namespace {
	case "kube-system",
		"istio-system",
		"cert-manager",
		"crossplane-system":
		score += defaultScore3
	}

	// ------------------------------------------------------------
	// Large change escalation
	// ------------------------------------------------------------

	if resource.ChangedLines > defaultChangedLength50 {
		score += defaultScore2
	}

	if resource.ChangedLines > defaultChangedLength200 {
		score += defaultScore3
	}

	// ------------------------------------------------------------
	// Final severity mapping
	// ------------------------------------------------------------

	switch {
	case score >= defaultScore9:
		return Critical

	case score >= defaultScore6:
		return High

	case score >= defaultScore3:
		return Medium

	default:
		return Low
	}
}

func buildCategoryMap() map[string]Category {
	result := make(map[string]Category)

	for category, resources := range categoryMappings {
		for resource := range resources {
			result[resource] = category
		}
	}

	return result
}

func resourceSet(resources ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(resources))

	for _, resource := range resources {
		result[resource] = struct{}{}
	}

	return result
}
