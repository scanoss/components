package dtoGetComponentVersion

type ComponentVersionsOutput struct {
	Component ComponentOutput `json:"component"`
}

type ComponentOutput struct {
	Component string             `json:"component"`
	Purl      string             `json:"purl"`
	Url       string             `json:"url"`
	Versions  []ComponentVersion `json:"versions"`
}

type ComponentVersion struct {
	Version  string             `json:"version"`
	Licenses []ComponentLicense `json:"licenses"`
}

type ComponentLicense struct {
	Name   string `json:"name"`
	SpdxId string `json:"spdx_id"`
	IsSpdx bool   `json:"is_spdx_approved"`
}
