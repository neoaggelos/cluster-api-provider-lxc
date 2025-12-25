package sync

// Manifest represents the images.yaml manifest used to sync images.
type Manifest struct {
	Images []Image `json:"images"`
}

// Image represents information for an image that must be synced.
type Image struct {
	ID       string   `json:"id"`
	Alias    []string `json:"alias"`
	Type     string   `json:"type"`
	Source   string   `json:"source"`
	Checksum string   `json:"checksum"`
}
