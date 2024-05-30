package models

type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type User struct {
	UserID   int      `json:"user_id"`
	UserName string   `json:"user_name"`
	Location Location `json:"location"`
}
