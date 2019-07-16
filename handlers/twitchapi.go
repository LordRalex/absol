package handlers

type TwitchApi struct {
	Data []TwitchApiData `json:"data"`
}

type TwitchApiData struct {
	Id string `json:"id"`
}
