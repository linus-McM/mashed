package advice

// AdviceMode represents a methodology advice perspective.
type AdviceMode struct {
	Name        string `json:"name" yaml:"name"`
	DisplayName string `json:"displayName" yaml:"displayName"`
	Icon        string `json:"icon" yaml:"icon"`
	Order       int    `json:"order" yaml:"order"`
	Body        string `json:"-" yaml:"-"`
	Source      string `json:"source"`
	FilePath    string `json:"-"`
}
