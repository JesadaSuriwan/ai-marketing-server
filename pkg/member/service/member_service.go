package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/repository"
)

type memberService struct {
	memberRepository repository.MemberRepository
}

func NewMemberService(memberRepository repository.MemberRepository) MemberService {
	return memberService{memberRepository}
}

func (s memberService) verifyOwnership(id, userId int) error {
	ok, err := s.memberRepository.BelongsToUser(id, userId)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}
	if !ok {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s memberService) GetAll(companyId int) (*MemberListResponse, error) {
	members, err := s.memberRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []MemberData{}
	for _, m := range members {
		data = append(data, MemberData{Id: m.Id, CompanyId: m.CompanyId, Email: m.Email, Name: m.Name, Role: m.Role, JoinedAt: m.JoinedAt})
	}

	return &MemberListResponse{Status: true, Desc: "Get members successful", Data: data}, nil
}

func (s memberService) Create(req CreateMemberRequest) (*SimpleResponse, error) {
	tx, err := s.memberRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	role := req.Role
	if role == "" {
		role = "Viewer"
	}

	m := repository.Member{CompanyId: req.CompanyId, Email: req.Email, Name: req.Name, Role: role}
	if _, err = s.memberRepository.Create(tx, m); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Member invited successfully"}, nil
}

func (s memberService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.memberRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.memberRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Member removed successfully"}, nil
}
