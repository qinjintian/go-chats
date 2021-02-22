package model

import "time"

type User struct {
	Id        int       `gorm:"primary_key" json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Activate  uint8     `json:"activate"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) TableName() string {
	return "gc_users"
}

func (u *User) Login() {

}
