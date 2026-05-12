package parser

type (
	ChangeType string
	Severity   string
	Category   string
)

const (
	Create ChangeType = "CREATE"
	Update ChangeType = "UPDATE"
	Delete ChangeType = "DELETE"
)

const (
	Low      Severity = "LOW"
	Medium   Severity = "MEDIUM"
	High     Severity = "HIGH"
	Critical Severity = "CRITICAL"
)

const (
	Workload   Category = "WORKLOAD"
	Networking Category = "NETWORKING"
	Security   Category = "SECURITY"
	Storage    Category = "STORAGE"
	Platform   Category = "PLATFORM"
	Config     Category = "CONFIG"
	Unknown    Category = "UNKNOWN"
)

type ResourceDiff struct {
	Kind         string     `yaml:"kind,omitempty"          json:"kind,omitempty"`
	Name         string     `yaml:"name,omitempty"          json:"name,omitempty"`
	Namespace    string     `yaml:"namespace,omitempty"     json:"namespace,omitempty"`
	ChangeType   ChangeType `yaml:"change_type,omitempty"   json:"change_type,omitempty"`
	Severity     Severity   `yaml:"severity,omitempty"      json:"severity,omitempty"`
	Category     Category   `yaml:"category,omitempty"      json:"category,omitempty"`
	Additions    int        `yaml:"additions,omitempty"     json:"additions,omitempty"`
	Deletions    int        `yaml:"deletions,omitempty"     json:"deletions,omitempty"`
	ChangedLines int        `yaml:"changed_lines,omitempty" json:"changed_lines,omitempty"`
}
