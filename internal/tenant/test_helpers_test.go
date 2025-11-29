package tenant

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dbInstance, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := dbInstance.AutoMigrate(
		&Tenant{},
		&TenantDomain{},
		&TenantMember{},
		&TenantIdentity{},
		&EmailProfile{},
		&SMSProfile{},
	); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return dbInstance
}

func writeBootstrapFile(t *testing.T, cfg BootstrapConfig) string {
	t.Helper()
	payload, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal bootstrap config: %v", err)
	}
	path := filepath.Join(t.TempDir(), "tenants.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("write bootstrap file: %v", err)
	}
	return path
}

func sampleBootstrapConfig() BootstrapConfig {
	return BootstrapConfig{
		Tenants: []BootstrapTenant{
			{
				ID:           "tenant-one",
				Slug:         "alpha",
				DisplayName:  "Alpha Corp",
				SupportEmail: "support@alpha.example",
				Status:       string(TenantStatusActive),
				Domains:      []string{"alpha.example", "portal.alpha.example"},
				Admins: []BootstrapMember{
					{Email: "admin@alpha.example", Role: "owner"},
					{Email: "viewer@alpha.example", Role: "viewer"},
				},
				Identity: BootstrapIdentity{
					GoogleClientID: "google-alpha",
					TAuthBaseURL:   "https://tauth.alpha.example",
				},
				EmailProfile: BootstrapEmailProfile{
					Host:        "smtp.alpha.example",
					Port:        587,
					Username:    "smtp-user",
					Password:    "smtp-pass",
					FromAddress: "noreply@alpha.example",
				},
				SMSProfile: &BootstrapSMSProfile{
					AccountSID: "AC123",
					AuthToken:  "sms-secret",
					FromNumber: "+10000000000",
				},
			},
		},
	}
}
