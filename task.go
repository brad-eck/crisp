package main

type Task struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"` // "Todo", "In Progress", "Done"
	Complete bool   `json:"complete"`
}