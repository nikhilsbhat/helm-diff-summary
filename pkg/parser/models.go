package parser

type ChangeType string

const (
	Create ChangeType = "CREATE"
	Update ChangeType = "UPDATE"
	Delete ChangeType = "DELETE"
)

type ResourceDiff struct {
	Kind         string         `yaml:"kind,omitempty"          json:"kind,omitempty"`
	Name         string         `yaml:"name,omitempty"          json:"name,omitempty"`
	Namespace    string         `yaml:"namespace,omitempty"     json:"namespace,omitempty"`
	ChangeType   ChangeType     `yaml:"change_type,omitempty"   json:"change_type,omitempty"`
	Additions    int            `yaml:"additions,omitempty"     json:"additions,omitempty"`
	Deletions    int            `yaml:"deletions,omitempty"     json:"deletions,omitempty"`
	ChangedLines int            `yaml:"changed_lines,omitempty" json:"changed_lines,omitempty"`
	FieldChanges map[string]int `yaml:"field_changes,omitempty" json:"field_changes,omitempty"`
}
