package api

import "context"

const ContextAdminKey ContextKey = "is_admin"

type ContextKey string

func isAdminUser(ctx context.Context) bool {
	isAdmin := ctx.Value(ContextAdminKey)

	if isAdmin == nil {
		return false
	}

	return isAdmin.(bool)
}
