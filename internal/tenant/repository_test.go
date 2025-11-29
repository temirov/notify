package tenant

import (
	"context"
	"testing"
)

func TestRepositoryResolveByHost(t *testing.T) {
	t.Helper()
	dbInstance := newTestDatabase(t)
	keeper := newTestSecretKeeper(t)
	configPath := writeBootstrapFile(t, sampleBootstrapConfig())
	if err := BootstrapFromFile(context.Background(), dbInstance, keeper, configPath); err != nil {
		t.Fatalf("bootstrap error: %v", err)
	}

	repo := NewRepository(dbInstance, keeper)
	runtimeCfg, err := repo.ResolveByHost(context.Background(), "portal.alpha.example")
	if err != nil {
		t.Fatalf("resolve host error: %v", err)
	}
	if runtimeCfg.Tenant.Slug != "alpha" {
		t.Fatalf("unexpected tenant slug %q", runtimeCfg.Tenant.Slug)
	}
	if runtimeCfg.Email.Username != "smtp-user" || runtimeCfg.Email.Password != "smtp-pass" {
		t.Fatalf("SMTP credentials not decrypted correctly")
	}
	if runtimeCfg.SMS == nil || runtimeCfg.SMS.AccountSID != "AC123" {
		t.Fatalf("expected SMS credentials")
	}
	if len(runtimeCfg.Admins) != 2 {
		t.Fatalf("expected 2 admins, got %d", len(runtimeCfg.Admins))
	}
}

func TestRepositoryListActiveTenants(t *testing.T) {
	t.Helper()
	dbInstance := newTestDatabase(t)
	keeper := newTestSecretKeeper(t)
	cfg := sampleBootstrapConfig()
	cfg.Tenants = append(cfg.Tenants, BootstrapTenant{
		ID:           "tenant-two",
		Slug:         "beta",
		DisplayName:  "Beta",
		SupportEmail: "support@beta.example",
		Status:       string(TenantStatusSuspended),
		Domains:      []string{"beta.example"},
		Admins:       []BootstrapMember{},
		Identity: BootstrapIdentity{
			GoogleClientID: "google-beta",
			TAuthBaseURL:   "https://tauth.beta.example",
		},
		EmailProfile: BootstrapEmailProfile{
			Host:        "smtp.beta.example",
			Port:        25,
			Username:    "beta-user",
			Password:    "beta-pass",
			FromAddress: "noreply@beta.example",
		},
	})
	configPath := writeBootstrapFile(t, cfg)
	if err := BootstrapFromFile(context.Background(), dbInstance, keeper, configPath); err != nil {
		t.Fatalf("bootstrap error: %v", err)
	}

	repo := NewRepository(dbInstance, keeper)
	tenants, err := repo.ListActiveTenants(context.Background())
	if err != nil {
		t.Fatalf("list active tenants error: %v", err)
	}
	if len(tenants) != 1 || tenants[0].Slug != "alpha" {
		t.Fatalf("expected only active tenant, got %+v", tenants)
	}
}

func TestRuntimeContextHelpers(t *testing.T) {
	t.Helper()
	cfg := RuntimeConfig{
		Tenant: Tenant{ID: "tenant-ctx", Slug: "ctx"},
	}
	ctx := context.Background()
	ctx = WithRuntime(ctx, cfg)
	result, ok := RuntimeFromContext(ctx)
	if !ok {
		t.Fatalf("expected runtime config")
	}
	if result.Tenant.ID != "tenant-ctx" {
		t.Fatalf("unexpected tenant in context")
	}
}
