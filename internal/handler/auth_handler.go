// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/handler/auth_handler.go
// Role: Interface adapters — auth gRPC handler
// Description: Translates the AuthService gRPC contract to/from the application service. The
// handler owns ONLY transport concerns: proto<->domain mapping, tenant resolution from metadata,
// and turning a model.AppError into the right gRPC status code. All business rules live in the
// service. Unimplemented RPCs (ResetPassword, SSO) fall through to the embedded base.

package handler

import (
	"context"
	"strconv"

	"neptune-pamm/github.com/ratheeshkumar25/internal/model"
	authv1 "neptune-pamm/github.com/ratheeshkumar25/internal/proto/auth/v1"
	"neptune-pamm/github.com/ratheeshkumar25/internal/services"

	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mdTenantID is the metadata header carrying the tenant for RPCs whose body has no tenant_id
// (Login). The gateway sets it from the authenticated host/subdomain — never from the client body.
const mdTenantID = "x-tenant-id"

// AuthHandler adapts the gRPC AuthService surface onto services.AuthService.
type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	svc services.AuthService
	log *slog.Logger
}

// NewAuthHandler builds the gRPC handler.
func NewAuthHandler(svc services.AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// Register creates a new self-service account.
func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	res, err := h.svc.Register(ctx, model.RegisterInput{
		TenantID:      req.GetTenantId(),
		Username:      req.GetUsername(),
		Email:         req.GetEmail(),
		Phone:         req.GetPhone(),
		Country:       req.GetCountry(),
		Password:      req.GetPassword(),
		Role:          roleFromProto(req.GetRole()),
		Currency:      req.GetCurrency(),
		ReferralCode:  req.GetReferralCode(),
		TermsAccepted: req.GetTermsAccepted(),
	})
	if err != nil {
		return nil, h.grpcError(err)
	}
	return &authv1.RegisterResponse{
		AccountId:                 res.AccountID,
		Username:                  res.Username,
		Role:                      roleToProto(res.Role),
		EmailVerificationRequired: res.EmailVerificationRequired,
	}, nil
}

// Login authenticates a PAMM principal. MT-mode login requires the external MT5 bridge.
func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req.GetMode() == authv1.LoginMode_LOGIN_MODE_MT {
		return nil, status.Error(codes.Unimplemented, "MT login requires the MT5 bridge")
	}

	tenantID, err := tenantFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	res, err := h.svc.Login(ctx, services.LoginInput{
		TenantID: tenantID,
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, h.grpcError(err)
	}
	return &authv1.LoginResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    res.ExpiresIn,
		AccountId:    res.AccountID,
		Role:         roleToProto(res.Role),
	}, nil
}

// Logout revokes the access token carried in the request body.
func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.svc.Logout(ctx, req.GetAccessToken()); err != nil {
		return nil, h.grpcError(err)
	}
	return &authv1.LogoutResponse{Success: true}, nil
}

// VerifyEmail consumes a verification token and unlocks login.
func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	if err := h.svc.VerifyEmail(ctx, req.GetToken()); err != nil {
		return nil, h.grpcError(err)
	}
	return &authv1.VerifyEmailResponse{Verified: true}, nil
}

// ListCurrencies returns the tenant's configured currencies.
func (h *AuthHandler) ListCurrencies(ctx context.Context, req *authv1.ListCurrenciesRequest) (*authv1.ListCurrenciesResponse, error) {
	currencies, err := h.svc.ListCurrencies(ctx, req.GetTenantId())
	if err != nil {
		return nil, h.grpcError(err)
	}
	out := make([]*authv1.Currency, 0, len(currencies))
	for i := range currencies {
		out = append(out, currencyToProto(currencies[i]))
	}
	return &authv1.ListCurrenciesResponse{Currencies: out}, nil
}

// --- mapping helpers --------------------------------------------------------------

// tenantFromMetadata reads and validates the x-tenant-id metadata header.
func tenantFromMetadata(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.InvalidArgument, "missing request metadata")
	}
	vals := md.Get(mdTenantID)
	if len(vals) == 0 {
		return 0, status.Errorf(codes.InvalidArgument, "missing %s header", mdTenantID)
	}
	id, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, status.Errorf(codes.InvalidArgument, "invalid %s header", mdTenantID)
	}
	return id, nil
}

// roleFromProto maps a proto role enum to the domain role. ROLE_UNSPECIFIED maps to an empty
// role, which the service rejects with a validation error.
func roleFromProto(r authv1.Role) model.Role {
	switch r {
	case authv1.Role_ROLE_ADMIN:
		return model.RoleAdmin
	case authv1.Role_ROLE_MONEY_MANAGER:
		return model.RoleMoneyManager
	case authv1.Role_ROLE_INVESTOR:
		return model.RoleInvestor
	default:
		return ""
	}
}

// roleToProto maps a domain role to the proto enum.
func roleToProto(r model.Role) authv1.Role {
	switch r {
	case model.RoleAdmin:
		return authv1.Role_ROLE_ADMIN
	case model.RoleMoneyManager:
		return authv1.Role_ROLE_MONEY_MANAGER
	case model.RoleInvestor:
		return authv1.Role_ROLE_INVESTOR
	default:
		return authv1.Role_ROLE_UNSPECIFIED
	}
}

// currencyToProto maps a domain currency to its wire form.
func currencyToProto(c model.Currency) *authv1.Currency {
	symbol := ""
	if c.Symbol != nil {
		symbol = *c.Symbol
	}
	return &authv1.Currency{
		Name:   c.Name,
		Digits: int32(c.Digits),
		Symbol: symbol,
	}
}

// grpcError converts a domain error into a gRPC status. Internal errors are logged with their
// cause but returned to the client as an opaque message so driver details never leak.
func (h *AuthHandler) grpcError(err error) error {
	appErr := model.FromError(err)
	if appErr.Code == model.ErrCodeInternal {
		h.log.Error("auth handler internal error", "err", err)
		return status.Error(codes.Internal, "internal server error")
	}
	return status.Error(grpcCode(appErr.Code), appErr.Message)
}

// grpcCode maps a domain error code to a gRPC status code.
func grpcCode(code model.ErrorCode) codes.Code {
	switch code {
	case model.ErrCodeData:
		return codes.InvalidArgument
	case model.ErrCodeUnauthorized:
		return codes.Unauthenticated
	case model.ErrCodeForbidden:
		return codes.PermissionDenied
	case model.ErrCodeNotFound:
		return codes.NotFound
	case model.ErrCodeConflict:
		return codes.AlreadyExists
	case model.ErrCodeRateLimited:
		return codes.ResourceExhausted
	case model.ErrCodeUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}
