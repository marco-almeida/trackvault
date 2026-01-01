package models

import "time"

type Playlist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsVirtual   bool   `json:"is_virtual"`
	IsPublic    bool   `json:"public"`
	Provider    string `json:"provider"`
	Owner       string `json:"owner"`
}

type Track struct {
	ID      string   `json:"id"`
	Artists []string `json:"artists"`
	Name    string   `json:"name"`
	Album   string   `json:"album"`
}

type User struct {
	ID          string
	DisplayName string
}

type BackupMetadata struct {
	Provider  string    `json:"provider"`
	Timestamp time.Time `json:"timestamp"`
	Result    string    `json:"status"`
}

type PlaylistWithTracksBackup struct {
	Playlist      Playlist `json:"playlist"`
	BackupSuccess bool     `json:"backup_success"`
	Tracks        []Track  `json:"tracks"`
}
