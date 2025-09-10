package services

import (
	"strconv"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// applyTenantFilter adds a tenant_id filter if the user is not a superadmin.
func applyTenantFilter(search utils.SearchParams, tenantID int, isSuper bool) utils.SearchParams {
	if isSuper {
		return search
	}
	if search.Filters == nil {
		search.Filters = make(map[string]string)
	}
	search.Filters["tenant_id"] = strconv.Itoa(tenantID)
	return search
}
