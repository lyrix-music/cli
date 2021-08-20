package routes

type PlayerChangeRequest struct {
	DbusId string `json:"dbus_id"`
}

type PlayerLyrics struct {
	Lyrics string `json:"lyrics,omitempty"`
	CJK    bool   `json:"cjk"`
}
