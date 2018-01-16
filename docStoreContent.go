package main

type DocStoreContent struct {
	Members []Member `json:"members"`
	Type    string   `json:"type"`
}

type Member struct {
	UUID string `json:"uuid"`
}
