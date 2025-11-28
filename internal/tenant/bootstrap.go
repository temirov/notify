package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BootstrapConfig defines the JSON layout for tenant provisioning.
type BootstrapConfig struct {
	Tenants []BootstrapTenant `json:"tenants"`
}

// BootstrapTenant declares per-tenant metadata.
type BootstrapTenant struct {
	ID           string                `json:"id"`
	Slug         string                `json:"slug"`
	DisplayName  string                `json:"displayName"`
	SupportEmail string                `json:"supportEmail"`
	Status       string                `json:"status"`
	Domains      []string              `json:"domains"`
	Admins       []BootstrapMember     `json:"admins"`
	Identity     BootstrapIdentity     `json:"identity"`
	EmailProfile BootstrapEmailProfile `json:"emailProfile"`
	SMSProfile   *BootstrapSMSProfile  `json:"smsProfile"`
}

// BootstrapMember captures admin membership entries.
type BootstrapMember struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// BootstrapIdentity holds GIS/TAuth metadata.
type BootstrapIdentity struct {
	GoogleClientID string `json:"googleClientId"`
	TAuthBaseURL   string `json:"tauthBaseUrl"`
}

// BootstrapEmailProfile defines SMTP credentials.
type BootstrapEmailProfile struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"fromAddress"`
}

// BootstrapSMSProfile defines Twilio credentials.
type BootstrapSMSProfile struct {
	AccountSID string `json:"accountSid"`
	AuthToken  string `json:"authToken"`
	FromNumber string `json:"fromNumber"`
}

// BootstrapFromFile loads tenants from a JSON file and upserts them.
func BootstrapFromFile(ctx context.Context, db *gorm.DB, keeper *SecretKeeper, path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("tenant bootstrap: read file: %w", err)
	}
	var cfg BootstrapConfig
	if err := json.Unmarshal(contents, &cfg); err != nil {
		return fmt.Errorf("tenant bootstrap: parse json: %w", err)
	}
	if len(cfg.Tenants) == 0 {
		return fmt.Errorf("tenant bootstrap: no tenants in %s", path)
	}
	for _, tenantSpec := range cfg.Tenants {
		if err := upsertTenant(ctx, db, keeper, tenantSpec); err != nil {
			return err
		}
	}
	return nil
}

func upsertTenant(ctx context.Context, db *gorm.DB, keeper *SecretKeeper, spec BootstrapTenant) error {
	if strings.TrimSpace(spec.ID) == "" {
		spec.ID = uuid.NewString()
	}
	if spec.Status == "" {
		spec.Status = string(TenantStatusActive)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenantModel := Tenant{
			ID:           spec.ID,
			Slug:         spec.Slug,
			DisplayName:  spec.DisplayName,
			SupportEmail: spec.SupportEmail,
			Status:       TenantStatus(spec.Status),
		}
		if err := tx.Clauses(clauseOnConflictUpdateAll()).
			Create(&tenantModel).Error; err != nil {
			return fmt.Errorf("tenant bootstrap: upsert tenant %s: %w", spec.Slug, err)
		}

		if err := tx.Where("tenant_id = ?", spec.ID).Delete(&TenantDomain{}).Error; err != nil {
			return err
		}
		for idx, host := range spec.Domains {
			domain := TenantDomain{
				TenantID:  spec.ID,
				Host:      strings.ToLower(host),
				IsDefault: idx == 0,
			}
			if err := tx.Create(&domain).Error; err != nil {
				return fmt.Errorf("tenant bootstrap: domain %s: %w", host, err)
			}
		}

		identity := TenantIdentity{
			TenantID:       spec.ID,
			GoogleClientID: spec.Identity.GoogleClientID,
			TAuthBaseURL:   spec.Identity.TAuthBaseURL,
		}
		if err := tx.Clauses(clauseOnConflictUpdateAll()).Create(&identity).Error; err != nil {
			return fmt.Errorf("tenant bootstrap: identity: %w", err)
		}

		if err := tx.Where("tenant_id = ?", spec.ID).Delete(&TenantMember{}).Error; err != nil {
			return err
		}
		for _, admin := range spec.Admins {
			member := TenantMember{
				TenantID: spec.ID,
				Email:    strings.ToLower(strings.TrimSpace(admin.Email)),
				Role:     strings.TrimSpace(admin.Role),
			}
			if member.Email == "" {
				continue
			}
			if member.Role == "" {
				member.Role = "admin"
			}
			if err := tx.Create(&member).Error; err != nil {
				return fmt.Errorf("tenant bootstrap: member %s: %w", member.Email, err)
			}
		}

		usernameCipher, err := keeper.Encrypt(spec.EmailProfile.Username)
		if err != nil {
			return err
		}
		passwordCipher, err := keeper.Encrypt(spec.EmailProfile.Password)
		if err != nil {
			return err
		}
		emailProfile := EmailProfile{
			ID:             uuid.NewString(),
			TenantID:       spec.ID,
			Host:           spec.EmailProfile.Host,
			Port:           spec.EmailProfile.Port,
			UsernameCipher: usernameCipher,
			PasswordCipher: passwordCipher,
			FromAddress:    spec.EmailProfile.FromAddress,
			IsDefault:      true,
		}
		if err := tx.Where("tenant_id = ?", spec.ID).Delete(&EmailProfile{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&emailProfile).Error; err != nil {
			return fmt.Errorf("tenant bootstrap: email profile: %w", err)
		}

		if spec.SMSProfile != nil {
			accountCipher, err := keeper.Encrypt(spec.SMSProfile.AccountSID)
			if err != nil {
				return err
			}
			tokenCipher, err := keeper.Encrypt(spec.SMSProfile.AuthToken)
			if err != nil {
				return err
			}
			smsProfile := SMSProfile{
				ID:               uuid.NewString(),
				TenantID:         spec.ID,
				AccountSIDCipher: accountCipher,
				AuthTokenCipher:  tokenCipher,
				FromNumber:       spec.SMSProfile.FromNumber,
				IsDefault:        true,
			}
			if err := tx.Where("tenant_id = ?", spec.ID).Delete(&SMSProfile{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&smsProfile).Error; err != nil {
				return fmt.Errorf("tenant bootstrap: sms profile: %w", err)
			}
		} else {
			if err := tx.Where("tenant_id = ?", spec.ID).Delete(&SMSProfile{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func clauseOnConflictUpdateAll() clause.Expression {
	return clause.OnConflict{
		UpdateAll: true,
	}
}
