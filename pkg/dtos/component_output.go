package dtos

type ComponentsSearchResults struct {
	Components []ComponentSearchResult `json:"components"`
}

type ComponentSearchResult struct {
	Component string
	Purl      string
	Url       string
}
