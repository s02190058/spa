package entity

import "time"

type Comment struct {
	ID      int       `json:"id"`
	Author  *User     `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
}
