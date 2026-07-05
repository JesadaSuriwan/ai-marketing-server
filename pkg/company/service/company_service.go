package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
)

type companyService struct {
	companyRepository repository.CompanyRepository
}

func NewCompanyService(companyRepository repository.CompanyRepository) CompanyService {
	return companyService{companyRepository}
}

func (s companyService) buildCompanyData(c repository.Company) CompanyData {
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
		Id:          c.Id,
		UserId:      c.UserId,
		Name:        c.Name,
		Initials:    c.Initials,
		LogoUrl:     c.LogoUrl,
		Industry:    c.Industry,
		Website:     c.Website,
		Location:    c.Location,
		Size:        c.Size,
		Founded:     c.Founded,
		Phone:       c.Phone,
		Email:       c.Email,
		About:       c.About,
		Plan:        c.Plan,
		Tags:        tagList,
		Markets:     marketList,
		SocialLinks: slDTO,
		Contacts:    contactList,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
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
		data = append(data, s.buildCompanyData(c))
	}

	return &CompanyListResponse{Status: true, Desc: "Get companies successful", Data: data}, nil
}

func (s companyService) Create(userId int, req CreateCompanyRequest) (*CompanyResponse, error) {
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
		UserId:   userId,
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
		Plan:     plan,
	}

	id, err := s.companyRepository.Create(tx, c)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	c.Id = id
	data := s.buildCompanyData(c)
	return &CompanyResponse{Status: true, Desc: "Company created successfully", Data: data}, nil
}

func (s companyService) GetById(id int) (*CompanyResponse, error) {
	c, err := s.companyRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("company not found")
	}

	data := s.buildCompanyData(*c)
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
