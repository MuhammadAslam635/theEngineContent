package services

import (
	"errors"
	"time"

	"backend-go/config"
	"backend-go/internal/dto"
	"backend-go/internal/helpers"
	"backend-go/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
	ForgetPassword(req dto.ForgetPasswordRequest) error
	ResetPassword(req dto.ResetPasswordRequest) error
}

type authService struct {
	repo repositories.AuthRepository
	cfg  *config.Config
}

func NewAuthService(repo repositories.AuthRepository, cfg *config.Config) AuthService {
	return &authService{repo: repo, cfg: cfg}
}

func (s *authService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := s.repo.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := helpers.GenerateToken(s.cfg.DBPassword, user.ID, user.Email, user.Username, user.UType, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	resp := &dto.LoginResponse{
		Token: token,
		User: dto.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			UType:    user.UType,
		},
	}

	return resp, nil
}

func (s *authService) ForgetPassword(req dto.ForgetPasswordRequest) error {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return errors.New("user not found")
	}

	otp, _ := helpers.GenerateNumericOTP()
	user.VerifyOTP = otp
	expires := time.Now().Add(15 * time.Minute)
	user.VerifyOTPExpires = &expires

	if err := s.repo.Update(user); err != nil {
		return err
	}

	return helpers.SendVerificationEmail(user.Email, user.Username, otp)
}

func (s *authService) ResetPassword(req dto.ResetPasswordRequest) error {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return errors.New("user not found")
	}

	if user.VerifyOTP != req.OTP || user.VerifyOTPExpires.Before(time.Now()) {
		return errors.New("invalid or expired OTP")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hashedPassword)
	user.VerifyOTP = ""
	user.VerifyOTPExpires = nil

	return s.repo.Update(user)
}
