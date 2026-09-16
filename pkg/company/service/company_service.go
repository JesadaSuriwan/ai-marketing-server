package service

import (
	"time"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
	"github.com/jmoiron/sqlx"
)

const contractDateFormat = "2006-01-02"

// resolveContractEndDate requires exactly one of durationMonths (a preset
// like 1/2/12, computed from today) or an explicit end date, and returns the
// final YYYY-MM-DD string to store.
func resolveContractEndDate(durationMonths *int, explicitDate *string) (string, error) {
	hasDuration := durationMonths != nil && *durationMonths > 0
	hasDate := explicitDate != nil && *explicitDate != ""

	if hasDuration == hasDate {
		return "", errs.NewBadRequestError("set exactly one of duration_months or contract_end_date")
	}
	if hasDuration {
		return time.Now().AddDate(0, *durationMonths, 0).Format(contractDateFormat), nil
	}
	if _, err := time.Parse(contractDateFormat, *explicitDate); err != nil {
		return "", errs.NewBadRequestError("contract_end_date must be in YYYY-MM-DD format")
	}
	return *explicitDate, nil
}

type companyService struct {
	companyRepository repository.CompanyRepository
	db                *sqlx.DB
}

func NewCompanyService(companyRepository repository.CompanyRepository, db *sqlx.DB) CompanyService {
	return companyService{companyRepository, db}
}

// buildCompanyData computes the caller's effective role for this company —
// "admin" if callerId owns it, otherwise c.MemberRole if GetAll's LEFT JOIN
// already populated it, otherwise looked up directly (GetById's path).
func (s companyService) buildCompanyData(c repository.Company, callerId int) CompanyData {
	role := permission.Admin
	if c.UserId != callerId {
		if c.MemberRole != nil {
			role = *c.MemberRole
		} else if r, err := permission.EffectiveRole(s.db, c.Id, callerId); err == nil {
			role = r
		} else {
			role = ""
		}
	}
	tags, _ := s.companyRepository.GetTags(c.Id)
	markets, _ := s.companyRepository.GetMarkets(c.Id)
	sl, _ := s.companyRepository.GetSocialLinks(c.Id)
	contacts, _ := s.companyRepository.GetKeyContacts(c.Id)

	tagList := []string{}
	for _, t := range tags {
		tagList = append(tagList, t.Tag)
	}

	marketList := []string{}
	for _, m := range markets {
		marketList = append(marketList, m.Market)
	}

	var slDTO *SocialLinksDTO
	if sl != nil {
		slDTO = &SocialLinksDTO{
			Linkedin: sl.Linkedin,
			Twitter:  sl.Twitter,
			Facebook: sl.Facebook,
		}
	}

	contactList := []ContactDTO{}
	for _, ct := range contacts {
		contactList = append(contactList, ContactDTO{
			Id:    ct.Id,
			Name:  ct.Name,
			Role:  ct.Role,
			Email: ct.Email,
			Phone: ct.Phone,
		})
	}

	return CompanyData{
		Id:              c.Id,
		UserId:          c.UserId,
		Name:            c.Name,
		Initials:        c.Initials,
		LogoUrl:         c.LogoUrl,
		Industry:        c.Industry,
		Website:         c.Website,
		Location:        c.Location,
		Size:            c.Size,
		Founded:         c.Founded,
		Phone:           c.Phone,
		Email:           c.Email,
		About:           c.About,
		Plan:            c.Plan,
		Tags:            tagList,
		Markets:         marketList,
		SocialLinks:     slDTO,
		Contacts:        contactList,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
		Role:            role,
		PromptLimit:     c.PromptLimit,
		ContractEndDate: c.ContractEndDate,
	}
}

func (s companyService) GetAll(userId int) (*CompanyListResponse, error) {
	companies, err := s.companyRepository.GetAll(userId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []CompanyData{}
	for _, c := range companies {
		data = append(data, s.buildCompanyData(c, userId))
	}

	return &CompanyListResponse{Status: true, Desc: "Get companies successful", Data: data}, nil
}

func (s companyService) Create(userId int, req CreateCompanyRequest) (*CompanyResponse, error) {
	contractEndDate, err := resolveContractEndDate(req.DurationMonths, req.ContractEndDate)
	if err != nil {
		return nil, err
	}

	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	plan := req.Plan
	if plan == "" {
		plan = "Free"
	}

	c := repository.Company{
		UserId:          userId,
		Name:            req.Name,
		Initials:        req.Initials,
		LogoUrl:         req.LogoUrl,
		Industry:        req.Industry,
		Website:         &req.Website,
		Location:        req.Location,
		Size:            req.Size,
		Founded:         req.Founded,
		Phone:           req.Phone,
		Email:           req.Email,
		About:           req.About,
		Plan:            plan,
		ContractEndDate: &contractEndDate,
	}

	id, err := s.companyRepository.Create(tx, c)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	// Claude is pricier than the other engines, so new workspaces start with
	// it off — Admin/Team Lead can flip it on directly, or a Specialist can
	// request it via the approval-email flow (pkg/claudeapproval).
	if err = s.companyRepository.SetAiEngine(tx, id, "claude", false); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	c.Id = id
	data := s.buildCompanyData(c, userId)
	return &CompanyResponse{Status: true, Desc: "Company created successfully", Data: data}, nil
}

func (s companyService) GetById(id, userId int) (*CompanyResponse, error) {
	c, err := s.companyRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("company not found")
	}

	data := s.buildCompanyData(*c, userId)
	return &CompanyResponse{Status: true, Desc: "Get company successful", Data: data}, nil
}

func (s companyService) Update(id int, req UpdateCompanyRequest) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	c := repository.Company{
		Id:       id,
		Name:     req.Name,
		Initials: req.Initials,
		LogoUrl:  req.LogoUrl,
		Industry: req.Industry,
		Website:  req.Website,
		Location: req.Location,
		Size:     req.Size,
		Founded:  req.Founded,
		Phone:    req.Phone,
		Email:    req.Email,
		About:    req.About,
		Plan:     req.Plan,
	}

	if err = s.companyRepository.Update(tx, c); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Company updated successfully"}, nil
}

func (s companyService) Delete(id int) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Company deleted successfully"}, nil
}

func (s companyService) AddTag(companyId int, tag string) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.AddTag(tx, companyId, tag); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Tag added successfully"}, nil
}

func (s companyService) DeleteTag(companyId int, tag string) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.DeleteTag(tx, companyId, tag); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Tag deleted successfully"}, nil
}

func (s companyService) AddMarket(companyId int, market string) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.AddMarket(tx, companyId, market); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Market added successfully"}, nil
}

func (s companyService) DeleteMarket(companyId int, market string) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.DeleteMarket(tx, companyId, market); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Market deleted successfully"}, nil
}

func (s companyService) UpdateSocialLinks(companyId int, req UpdateSocialLinksRequest) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.UpsertSocialLinks(tx, companyId, req.Linkedin, req.Twitter, req.Facebook); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Social links updated successfully"}, nil
}

func (s companyService) AddContact(companyId int, req AddContactRequest) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	contact := repository.CompanyKeyContact{
		CompanyId: companyId,
		Name:      req.Name,
		Role:      req.Role,
		Email:     req.Email,
		Phone:     req.Phone,
	}

	if _, err = s.companyRepository.AddKeyContact(tx, contact); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Contact added successfully"}, nil
}

func (s companyService) DeleteContact(companyId, contactId int) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.DeleteKeyContact(tx, companyId, contactId); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Contact deleted successfully"}, nil
}

func (s companyService) GetAiEngines(companyId int) (*AiEngineListResponse, error) {
	rows, err := s.companyRepository.GetAiEngines(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	enabled := map[string]bool{}
	for _, r := range rows {
		enabled[r.Platform] = r.Enabled
	}

	data := []AiEngineData{}
	for _, platform := range AllAiEngines {
		e, ok := enabled[platform]
		if !ok {
			e = true // no row yet means the engine defaults to enabled
		}
		data = append(data, AiEngineData{Platform: platform, Label: aiEngineLabels[platform], Enabled: e})
	}

	return &AiEngineListResponse{Status: true, Desc: "Get AI engines successful", Data: data}, nil
}

func (s companyService) UpdateAiEngine(companyId int, platform string, enabled bool) (*SimpleResponse, error) {
	if _, ok := aiEngineLabels[platform]; !ok {
		return nil, errs.NewBadRequestError("unknown AI engine")
	}

	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.SetAiEngine(tx, companyId, platform, enabled); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "AI engine updated successfully"}, nil
}

func (s companyService) GetNotificationPrefs(companyId int) (*NotificationPrefsResponse, error) {
	p, err := s.companyRepository.GetNotificationPrefs(companyId)
	if err != nil {
		// No row yet — defaults match the table's column defaults, so a
		// company that never touched this page sees the same values it
		// would get if a row existed.
		return &NotificationPrefsResponse{
			Status: true, Desc: "Get notification preferences successful",
			Data: NotificationPrefsData{EmailReports: true, VisibilityAlerts: true, MemberActivity: false, WeeklyDigest: true},
		}, nil
	}

	return &NotificationPrefsResponse{
		Status: true, Desc: "Get notification preferences successful",
		Data: NotificationPrefsData{
			EmailReports: p.EmailReports, VisibilityAlerts: p.VisibilityAlerts,
			MemberActivity: p.MemberActivity, WeeklyDigest: p.WeeklyDigest,
		},
	}, nil
}

func (s companyService) UpdateNotificationPrefs(companyId int, req UpdateNotificationPrefsRequest) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	prefs := repository.CompanyNotificationPrefs{
		CompanyId: companyId, EmailReports: req.EmailReports, VisibilityAlerts: req.VisibilityAlerts,
		MemberActivity: req.MemberActivity, WeeklyDigest: req.WeeklyDigest,
	}
	if err = s.companyRepository.UpsertNotificationPrefs(tx, companyId, prefs); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Notification preferences updated successfully"}, nil
}

func (s companyService) UpdatePromptLimit(companyId int, req UpdatePromptLimitRequest) (*SimpleResponse, error) {
	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.SetPromptLimit(tx, companyId, req.Limit); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Prompt limit updated successfully"}, nil
}

func (s companyService) UpdateContract(companyId int, req UpdateContractRequest) (*SimpleResponse, error) {
	contractEndDate, err := resolveContractEndDate(req.DurationMonths, req.ContractEndDate)
	if err != nil {
		return nil, err
	}

	tx, err := s.companyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.SetContractEndDate(tx, companyId, contractEndDate); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Contract updated successfully"}, nil
}
