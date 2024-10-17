package models

type Language struct {
	Bytes int `json:"bytes"`
}

type Repo struct {
	FullName   *string             `json:"full_name,omitempty"`
	Owner      *string             `json:"owner,omitempty"`
	Repository *string             `json:"repository,omitempty"`
	Languages  map[string]Language `json:"languages,omitempty"`
}
