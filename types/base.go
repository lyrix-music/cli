package types

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
	Playing bool
	Track   string `json:"track"`
	Artist  string `json:"artist"`
}
