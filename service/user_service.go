package service

import (
	"context"
	"math"

	"go.uber.org/zap"

	"user-microservice-golang/config"
	apperrors "user-microservice-golang/errors"
	"user-microservice-golang/interfaces"
	"user-microservice-golang/model"
	"user-microservice-golang/utils"
)

// userService is the concrete implementation of interfaces.UserService
type userService struct {
	repo   interfaces.UserRepository
	cfg    *config.AppConfig
	logger *zap.Logger
}

// NewUserService constructs a userService
func NewUserService(repo interfaces.UserRepository, cfg *config.AppConfig, logger *zap.Logger) interfaces.UserService {
	return &userService{repo: repo, cfg: cfg, logger: logger}
}

// ─── Auth ────────────────────────────────────────────────────────────────────

func (s *userService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrEmailAlreadyExists
	}

	hashed, err := utils.HashPassword(req.Password, s.cfg.BcryptCost)
	if err != nil {
		s.logger.Error("hash password failed", zap.Error(err))
		return nil, apperrors.ErrInternal
	}

	user := &model.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashed,
		Phone:     req.Phone,
		Role:      model.RoleUser,
		Status:    model.StatusActive,
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(created)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	resp := model.ToUserResponse(created)
	return &model.AuthResponse{Token: token, User: resp}, nil
}

func (s *userService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		// mask not-found as invalid credentials (security best practice)
		return nil, apperrors.ErrInvalidCredentials
	}

	if user.Status == model.StatusBanned {
		return nil, apperrors.ErrForbidden
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	resp := model.ToUserResponse(user)
	return &model.AuthResponse{Token: token, User: resp}, nil
}

// ─── Queries ─────────────────────────────────────────────────────────────────

func (s *userService) GetByID(ctx context.Context, id string) (*model.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := model.ToUserResponse(user)
	return &resp, nil
}

func (s *userService) GetAll(ctx context.Context, page, limit int) (*model.PaginatedUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := s.repo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	data := make([]model.UserResponse, 0, len(users))
	for _, u := range users {
		data = append(data, model.ToUserResponse(u))
	}

	return &model.PaginatedUsersResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int64(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

// ─── Mutations ───────────────────────────────────────────────────────────────

func (s *userService) UpdateProfile(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.UserResponse, error) {
	updates := map[string]interface{}{}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}

	if len(updates) == 0 {
		return s.GetByID(ctx, id)
	}

	updated, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	resp := model.ToUserResponse(updated)
	return &resp, nil
}

func (s *userService) UpdatePassword(ctx context.Context, id string, req *model.UpdatePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(user.Password, req.CurrentPassword) {
		return apperrors.ErrInvalidCredentials
	}

	hashed, err := utils.HashPassword(req.NewPassword, s.cfg.BcryptCost)
	if err != nil {
		s.logger.Error("hash password failed", zap.Error(err))
		return apperrors.ErrInternal
	}

	return s.repo.UpdatePassword(ctx, id, hashed)
}

func (s *userService) UpdateStatus(ctx context.Context, id string, req *model.UpdateStatusRequest) (*model.UserResponse, error) {
	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *userService) generateToken(user *model.User) (string, error) {
	return utils.GenerateToken(
		user.ID.Hex(),
		user.Email,
		string(user.Role),
		s.cfg.JWTSecret,
		s.cfg.JWTExpiryHours,
	)
}
