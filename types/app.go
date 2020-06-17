package types

type App struct {
	UID      string   `json:"uid"`
	Manifest Manifest `json:"manifest"`
	IsGit    bool     `json:"git"`
}
