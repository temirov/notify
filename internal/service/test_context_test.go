package service

import (
	"context"

	"github.com/temirov/pinguin/internal/tenant"
)

const testTenantID = "tenant-service"

func tenantContext() context.Context {
	return tenant.WithRuntime(context.Background(), tenant.RuntimeConfig{
		Tenant: tenant.Tenant{
			ID:   testTenantID,
			Slug: "slug-" + testTenantID,
		},
	})
}
