package tenant

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// RuntimeConfig aggregates tenant data required at runtime.
type RuntimeConfig struct {
	Tenant   Tenant
	Identity TenantIdentity
	Admins   map[string]string
	Email    EmailCredentials
	SMS      *SMSCredentials
}

// EmailCredentials exposes decrypted SMTP settings.
type EmailCredentials struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
}

// SMSCredentials exposes decrypted Twilio settings.
type SMSCredentials struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

// Repository exposes tenant lookups.
type Repository struct {
	db     *gorm.DB
	keeper *SecretKeeper
}

// NewRepository constructs a repository.
func NewRepository(db *gorm.DB, keeper *SecretKeeper) *Repository {
	return &Repository{db: db, keeper: keeper}
}

// ResolveByHost returns the tenant associated with the provided host.
func (repo *Repository) ResolveByHost(ctx context.Context, host string) (RuntimeConfig, error) {
	normalized := normalizeHost(host)
	if normalized == "" {
		return RuntimeConfig{}, fmt.Errorf("tenant resolve: empty host")
	}
	var domain TenantDomain
	if err := repo.db.WithContext(ctx).Where("host = ?", normalized).First(&domain).Error; err != nil {
		return RuntimeConfig{}, fmt.Errorf("tenant resolve: domain %s: %w", normalized, err)
	}
	return repo.runtimeConfig(ctx, domain.TenantID)
}

// ResolveByID fetches tenant runtime config by id.
func (repo *Repository) ResolveByID(ctx context.Context, tenantID string) (RuntimeConfig, error) {
	return repo.runtimeConfig(ctx, tenantID)
}

// ListActiveTenants returns active tenant rows.
func (repo *Repository) ListActiveTenants(ctx context.Context) ([]Tenant, error) {
	var tenants []Tenant
	if err := repo.db.WithContext(ctx).
		Where("status = ?", TenantStatusActive).
		Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("tenant list: %w", err)
	}
	return tenants, nil
}

func (repo *Repository) runtimeConfig(ctx context.Context, tenantID string) (RuntimeConfig, error) {
	var tenantModel Tenant
	if err := repo.db.WithContext(ctx).Where("id = ?", tenantID).First(&tenantModel).Error; err != nil {
		return RuntimeConfig{}, fmt.Errorf("tenant runtime: tenant %s: %w", tenantID, err)
	}
	var identity TenantIdentity
	if err := repo.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&identity).Error; err != nil {
		return RuntimeConfig{}, fmt.Errorf("tenant runtime: identity: %w", err)
	}
	var emailProfile EmailProfile
	if err := repo.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Where("is_default = ?", true).First(&emailProfile).Error; err != nil {
		return RuntimeConfig{}, fmt.Errorf("tenant runtime: email profile: %w", err)
	}
	var smsPtr *SMSCredentials
	var smsProfile SMSProfile
	if err := repo.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where("is_default = ?", true).
		First(&smsProfile).Error; err == nil {
		accountSID, err := repo.keeper.Decrypt(smsProfile.AccountSIDCipher)
		if err != nil {
			return RuntimeConfig{}, err
		}
		authToken, err := repo.keeper.Decrypt(smsProfile.AuthTokenCipher)
		if err != nil {
			return RuntimeConfig{}, err
		}
		smsPtr = &SMSCredentials{
			AccountSID: accountSID,
			AuthToken:  authToken,
			FromNumber: smsProfile.FromNumber,
		}
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return RuntimeConfig{}, fmt.Errorf("tenant runtime: sms profile: %w", err)
	}
	admins := make(map[string]string)
	var members []TenantMember
	if err := repo.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&members).Error; err != nil {
		return RuntimeConfig{}, fmt.Errorf("tenant runtime: members: %w", err)
	}
	for _, member := range members {
		admins[normalizeEmail(member.Email)] = member.Role
	}
	username, err := repo.keeper.Decrypt(emailProfile.UsernameCipher)
	if err != nil {
		return RuntimeConfig{}, err
	}
	password, err := repo.keeper.Decrypt(emailProfile.PasswordCipher)
	if err != nil {
		return RuntimeConfig{}, err
	}
	return RuntimeConfig{
		Tenant:   tenantModel,
		Identity: identity,
		Admins:   admins,
		Email: EmailCredentials{
			Host:        emailProfile.Host,
			Port:        emailProfile.Port,
			Username:    username,
			Password:    password,
			FromAddress: emailProfile.FromAddress,
		},
		SMS: smsPtr,
	}, nil
}

func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		return parts[0]
	}
	return host
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
