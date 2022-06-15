package dtoSearchComponent

type ComponentsSearchOutput struct {
	Components []ComponentSearchOutput `json:"components"`
}

type ComponentSearchOutput struct {
	Component string
	Purl      string
	Url       string
}
