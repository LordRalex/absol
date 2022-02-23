package twitch

type TwitchApi struct {
	Data []TwitchApiData `json:"data"`
}

type TwitchApiData struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
	Login       string `json:"login"`
}
