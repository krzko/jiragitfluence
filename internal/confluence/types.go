package confluence

// Page represents a Confluence page
type Page struct {
	ID      string  `json:"id,omitempty"`
	Type    string  `json:"type"`
	Title   string  `json:"title"`
	Space   Space   `json:"space"`
	Body    Body    `json:"body"`
	Version Version `json:"version,omitempty"`
}

// Space represents a Confluence space
type Space struct {
	Key string `json:"key"`
}

// Body represents the content body of a Confluence page
type Body struct {
	Storage Storage `json:"storage"`
}

// Storage represents the storage format of Confluence content
type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// Version represents the version information of a Confluence page
type Version struct {
	Number    int    `json:"number"`
	Message   string `json:"message,omitempty"`
	MinorEdit bool   `json:"minorEdit"`
}

// SearchResult represents the result of a Confluence content search
type SearchResult struct {
	Results []struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Title   string `json:"title"`
		Version struct {
			Number int `json:"number"`
		} `json:"version,omitempty"`
		// Additional fields from the API response
		History struct {
			CreatedBy struct {
				DisplayName string `json:"displayName"`
				Username    string `json:"username"`
			} `json:"createdBy,omitempty"`
			CreatedDate string `json:"createdDate,omitempty"`
		} `json:"history,omitempty"`
		Links struct {
			Self string `json:"self"`
		} `json:"_links,omitempty"`
	} `json:"results"`
	Size  int `json:"size"`
	Limit int `json:"limit"`
	Start int `json:"start"`
}
