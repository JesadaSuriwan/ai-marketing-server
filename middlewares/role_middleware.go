package middlewares

import (
	"net/http"

	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type RoleMiddleware interface {
	// RequireRole must run after a CompanyMiddleware check has already set
	// "companyId" in context. Aborts 403 unless the caller's effective role
	// for that company is one of the given roles.
	RequireRole(roles ...string) gin.HandlerFunc

	// CanCreateWorkspace aborts 403 unless the authenticated user is allowed
	// to create a new company — see permission.CanCreateWorkspace.
	CanCreateWorkspace(c *gin.Context)
}

type roleMiddleware struct {
	db *sqlx.DB
}

func NewRoleMiddleware(db *sqlx.DB) RoleMiddleware {
	return roleMiddleware{db}
}

func (m roleMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ok := c.Get("userId")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
			return
		}
		companyIdVal, ok := c.Get("companyId")
		if !ok {
			// Programmer error: RequireRole must sit after a CompanyMiddleware
			// check in the route chain.
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": false, "desc": "role check misconfigured"})
			return
		}

		role, err := permission.EffectiveRole(m.db, companyIdVal.(int), userId.(int))
		if err != nil || !permission.Allowed(role, roles...) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
			return
		}
		c.Next()
	}
}

func (m roleMiddleware) CanCreateWorkspace(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	allowed, err := permission.CanCreateWorkspace(m.db, userId.(int))
	if err != nil || !allowed {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "only an Admin or Team Lead can create a new workspace"})
		return
	}
	c.Next()
}
