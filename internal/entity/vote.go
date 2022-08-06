package entity

type Vote struct {
	UserID int `json:"user"`
	Vote   int `json:"vote"`
}
