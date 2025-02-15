package models

type User struct {
	ID            int     `json:"id"`
	CreatedAt     string  `json:"created_at"`
	ChatID        string  `json:"chat_id"`
	UserID        string  `json:"user_id"`
	From          string  `json:"from"`
	Username      string  `json:"username"`
	ActiveCommand *string `json:"active_command"`
	HashPass      string  `json:"hash_pass"`
	Email         string  `json:"email"`
	IP            string  `json:"ip"`
}
