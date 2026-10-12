package port

// ScenarioRequest represents an HTTP request to execute as part of a scenario step.
type ScenarioRequest struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     string            `json:"body,omitempty"`
	Curl     string            `json:"curl"`
	Runnable bool              `json:"runnable"`
}

// ScenarioExpect represents the expected outcome under a specific configuration.
type ScenarioExpect struct {
	Config string `json:"config"`
	Result string `json:"result"`
}

// Scenario represents a tutorial scenario for CMC CDN, WAAP, or Shop flows.
type Scenario struct {
	ID          string            `json:"id"`
	Group       string            `json:"group"` // 'cdn' | 'waap' | 'shop'
	Title       string            `json:"title"`
	Goal        string            `json:"goal"`
	PortalSteps []string          `json:"portal_steps"`
	Requests    []ScenarioRequest `json:"requests"`
	Expect      []ScenarioExpect  `json:"expect"`
	Notes       string            `json:"notes"`
	DocsURL     string            `json:"docs_url"`
	Priority    string            `json:"priority"` // 'P0' | 'P1'
}

// ScenarioRepository provides access to tutorial scenarios.
type ScenarioRepository interface {
	ListScenarios() ([]Scenario, error)
	GetScenario(id string) (*Scenario, error)
}
