package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jmoiron/sqlx"
)

type CompanyMiddleware interface {
	// OwnerByQuery verifies company_id query param belongs to the authenticated
	// user, either as the owner or an accepted team member.
	// Use on list/aggregate endpoints: GET /dashboard?company_id=X
	OwnerByQuery(c *gin.Context)

	// OwnerByParam verifies the :id URL param is a company the authenticated
	// user owns or is an accepted member of.
	// Use on company router: GET/PUT /company/:id
	OwnerByParam(c *gin.Context)

	// OwnerOnlyByParam is the strict, owner-only variant of OwnerByParam —
	// team members do NOT pass this check. Reserve for irreversible/
	// account-level actions (e.g. DELETE /company/:id) that shouldn't be
	// available to every member regardless of role.
	OwnerOnlyByParam(c *gin.Context)

	// OwnerByBrandId verifies brand_id query param belongs to a company the
	// authenticated user owns or is an accepted member of.
	// Use on visibility list endpoints: GET /visibility?brand_id=X
	OwnerByBrandId(c *gin.Context)

	// OwnerByPromptId verifies prompt_id query param belongs to a company the
	// authenticated user owns or is an accepted member of.
	// Use on citation list endpoints: GET /citation?prompt_id=X
	OwnerByPromptId(c *gin.Context)

	// OwnerByBodyCompanyId verifies company_id in the JSON body belongs to a
	// company the authenticated user owns or is an accepted member of.
	// Use on create endpoints where company_id is in the request body.
	// The body is cached so the handler can re-read it with ShouldBindBodyWith.
	OwnerByBodyCompanyId(c *gin.Context)

	// OwnerByBodyBrandId verifies brand_id in the JSON body belongs to a
	// company the authenticated user owns or is an accepted member of.
	// Use on visibility create endpoint.
	OwnerByBodyBrandId(c *gin.Context)

	// OwnerByBodyPromptId verifies prompt_id in the JSON body belongs to a
	// company the authenticated user owns or is an accepted member of.
	// Use on citation create endpoint.
	OwnerByBodyPromptId(c *gin.Context)
}

type companyMiddleware struct {
	db *sqlx.DB
}

func NewCompanyMiddleware(db *sqlx.DB) CompanyMiddleware {
	return companyMiddleware{db}
}

// memberOrOwnerClause is appended to every access check below: true if the
// user directly owns the company, OR has an accepted (status='active')
// company_members row for it. $1 is always the company id in scope for the
// subquery, $%d is the userId placeholder position (varies per query).
func memberOrOwnerClause(companyIdExpr string, userIdPlaceholder string) string {
	return "(c.user_id = " + userIdPlaceholder + " OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = " + companyIdExpr + " AND cm.user_id = " + userIdPlaceholder + " AND cm.status = 'active'))"
}

func (m companyMiddleware) OwnerByQuery(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil || companyId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid company_id"})
		return
	}

	if !m.isCompanyOwner(companyId, userId.(int)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("companyId", companyId)
	c.Next()
}

func (m companyMiddleware) OwnerByParam(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	companyId, err := strconv.Atoi(c.Param("id"))
	if err != nil || companyId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid id"})
		return
	}

	if !m.isCompanyOwner(companyId, userId.(int)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("companyId", companyId)
	c.Next()
}

func (m companyMiddleware) OwnerOnlyByParam(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	companyId, err := strconv.Atoi(c.Param("id"))
	if err != nil || companyId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid id"})
		return
	}

	var owned bool
	m.db.Get(&owned, `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1 AND user_id = $2)`, companyId, userId.(int))
	if !owned {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("companyId", companyId)
	c.Next()
}

func (m companyMiddleware) OwnerByBrandId(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	brandId, err := strconv.Atoi(c.Query("brand_id"))
	if err != nil || brandId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid brand_id"})
		return
	}

	var owned bool
	err = m.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM brands b
			JOIN companies c ON b.company_id = c.id
			WHERE b.id = $1 AND `+memberOrOwnerClause("c.id", "$2")+`
		)`, brandId, userId.(int))
	if err != nil || !owned {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("brandId", brandId)
	c.Next()
}

func (m companyMiddleware) OwnerByPromptId(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	promptId, err := strconv.Atoi(c.Query("prompt_id"))
	if err != nil || promptId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid prompt_id"})
		return
	}

	var owned bool
	err = m.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM prompts p
			JOIN companies c ON p.company_id = c.id
			WHERE p.id = $1 AND `+memberOrOwnerClause("c.id", "$2")+`
		)`, promptId, userId.(int))
	if err != nil || !owned {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("promptId", promptId)
	c.Next()
}

func (m companyMiddleware) OwnerByBodyCompanyId(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	var body struct {
		CompanyId int `json:"company_id,string"`
	}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil || body.CompanyId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid company_id"})
		return
	}

	if !m.isCompanyOwner(body.CompanyId, userId.(int)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("companyId", body.CompanyId)
	c.Next()
}

func (m companyMiddleware) OwnerByBodyBrandId(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	var body struct {
		BrandId int `json:"brand_id,string"`
	}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil || body.BrandId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid brand_id"})
		return
	}

	var owned bool
	m.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM brands b
			JOIN companies c ON b.company_id = c.id
			WHERE b.id = $1 AND `+memberOrOwnerClause("c.id", "$2")+`
		)`, body.BrandId, userId.(int))
	if !owned {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("brandId", body.BrandId)
	c.Next()
}

func (m companyMiddleware) OwnerByBodyPromptId(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": false, "desc": "unauthorized"})
		return
	}

	var body struct {
		PromptId int `json:"prompt_id,string"`
	}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil || body.PromptId <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "desc": "invalid prompt_id"})
		return
	}

	var owned bool
	m.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM prompts p
			JOIN companies c ON p.company_id = c.id
			WHERE p.id = $1 AND `+memberOrOwnerClause("c.id", "$2")+`
		)`, body.PromptId, userId.(int))
	if !owned {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": false, "desc": "access denied"})
		return
	}

	c.Set("promptId", body.PromptId)
	c.Next()
}

func (m companyMiddleware) isCompanyOwner(companyId, userId int) bool {
	var owned bool
	m.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM companies c
			WHERE c.id = $1 AND `+memberOrOwnerClause("c.id", "$2")+`
		)`, companyId, userId)
	return owned
}