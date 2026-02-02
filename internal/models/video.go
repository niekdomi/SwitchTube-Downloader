package models

// Video represents a Video.
type Video struct {
	ID      string `json:"id"`      // The video ID
	Title   string `json:"title"`   // The video title
	Episode string `json:"episode"` // The episode number
}
