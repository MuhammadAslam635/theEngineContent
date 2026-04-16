package dto

type ResetPasswordRequest struct {
	OTP             string `json:"otp" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}
