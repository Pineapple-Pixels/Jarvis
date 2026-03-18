package domain

import "encoding/json"

// Figma path/query parameter constants.
const (
	PathParamFileKey   = "file_key"
	PathParamProjectID = "project_id"

	QueryParamNodeIDs = "ids"
	QueryParamFormat  = "format"
	QueryParamScale   = "scale"
)

// FigmaFile represents a Figma file's metadata.
type FigmaFile struct {
	Name         string `json:"name"`
	LastModified string `json:"last_modified"`
	ThumbnailURL string `json:"thumbnail_url"`
	Version      string `json:"version"`
}

// FigmaComponent represents a Figma component.
type FigmaComponent struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FigmaNodeDetail represents a node with its document tree.
type FigmaNodeDetail struct {
	Document   json.RawMessage        `json:"document"`
	Components map[string]FigmaComponent `json:"components,omitempty"`
}

// FigmaImage maps a node to its rendered image URL.
type FigmaImage struct {
	NodeID   string `json:"node_id"`
	ImageURL string `json:"image_url"`
}

// FigmaComment represents a comment on a Figma file.
type FigmaComment struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	User      string `json:"user"`
}

// FigmaProjectFile represents a file in a Figma project.
type FigmaProjectFile struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnail_url"`
	LastModified string `json:"last_modified"`
}

// FigmaFileResponse is the response for getting a file.
type FigmaFileResponse struct {
	Success bool       `json:"success"`
	File    *FigmaFile `json:"file,omitempty"`
	Error   string     `json:"error,omitempty"`
}

// FigmaNodesResponse is the response for getting nodes.
type FigmaNodesResponse struct {
	Success bool                       `json:"success"`
	Nodes   map[string]FigmaNodeDetail `json:"nodes,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

// FigmaImagesResponse is the response for rendering images.
type FigmaImagesResponse struct {
	Success bool         `json:"success"`
	Images  []FigmaImage `json:"images,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// FigmaCommentsResponse is the response for listing comments.
type FigmaCommentsResponse struct {
	Success  bool           `json:"success"`
	Comments []FigmaComment `json:"comments,omitempty"`
	Error    string         `json:"error,omitempty"`
}

// FigmaProjectFilesResponse is the response for listing project files.
type FigmaProjectFilesResponse struct {
	Success bool               `json:"success"`
	Files   []FigmaProjectFile `json:"files,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// FigmaComponentsResponse is the response for listing components.
type FigmaComponentsResponse struct {
	Success    bool             `json:"success"`
	Components []FigmaComponent `json:"components,omitempty"`
	Error      string           `json:"error,omitempty"`
}
