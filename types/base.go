package types

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRegisterRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TelegramId int    `json:"telegram_id"`
}

type UserAuthGrant struct {
	Token string `json:"token"`
}

type UserInstance struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Host     string `json:"host"`
}

type SongMeta struct {
	Playing    bool
	Track      string `json:"track"`
	Artist     string `json:"artist"`
	Source     string `json:"source,omitempty"`
	Url        string `json:"url,omitempty"`
	Scrobble   bool   `json:"scrobble,omitempty"`
	AlbumArt   string `json:"album_art,omitempty"`
	Mbid       string `json:"mbid,omitempty"`
	ArtistMbid string `json:"artist_mbid,omitempty"`
	Position int64
	IsRepeat bool `json:"is_repeat"`
}
