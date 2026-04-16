package dto

type ForgetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgetPasswordResponse struct {
	Message string `json:"message"`
}
