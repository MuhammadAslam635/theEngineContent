package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserDTO struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	UType    string `json:"utype"`
}

type LoginResponse struct {
	Token        string  `json:"token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
}
