package domain

// SpotifyTrack represents a Spotify track.
type SpotifyTrack struct {
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	URL       string `json:"url"`
	IsPlaying bool   `json:"is_playing"`
}

// SpotifyTrackResponse is the response for the currently playing track.
type SpotifyTrackResponse struct {
	Success bool          `json:"success"`
	Track   *SpotifyTrack `json:"track,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// SpotifyActionResponse is the response for play/pause/next actions.
type SpotifyActionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
