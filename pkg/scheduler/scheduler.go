package scheduler

import (
	"fmt"
	"time"

	"github.com/ai-marketing/ai-marketing-server/logs"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	notificationRepository "github.com/ai-marketing/ai-marketing-server/pkg/notification/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptrun/service"
	"github.com/robfig/cron/v3"
)

// Scheduler runs every tracked prompt (across every company) on a cron
// cadence, reusing the same run + citation-extraction pipeline as the
// manual "Run Prompt" button and the "run all" endpoint. It also sweeps for
// contracts about to expire (see contract_reminder.go) on the same cadence.
type Scheduler struct {
	cron                   *cron.Cron
	promptRunService       service.PromptRunService
	promptRepository       repository.PromptRepository
	companyRepository      companyRepository.CompanyRepository
	notificationRepository notificationRepository.NotificationRepository
}

// bangkokLocation pins the schedule to Thailand time regardless of the
// host's own system timezone (e.g. a UTC-configured VPS), so cron_expression
// in config always means "that time in Bangkok" rather than silently
// depending on wherever this happens to be deployed.
var bangkokLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		logs.Error(fmt.Errorf("scheduler: failed to load Asia/Bangkok location, falling back to UTC: %w", err))
		return time.UTC
	}
	return loc
}()

func New(
	promptRunService service.PromptRunService,
	promptRepository repository.PromptRepository,
	companyRepository companyRepository.CompanyRepository,
	notificationRepository notificationRepository.NotificationRepository,
) *Scheduler {
	return &Scheduler{
		cron:                   cron.New(cron.WithLocation(bangkokLocation)),
		promptRunService:       promptRunService,
		promptRepository:       promptRepository,
		companyRepository:      companyRepository,
		notificationRepository: notificationRepository,
	}
}

// Start schedules the prompt sweep and the contract-reminder sweep on the
// same cronExpr (standard 5-field cron syntax) and begins running both in
// the background.
func (s *Scheduler) Start(cronExpr string) error {
	if _, err := s.cron.AddFunc(cronExpr, s.runAll); err != nil {
		return err
	}
	if _, err := s.cron.AddFunc(cronExpr, s.checkContractReminders); err != nil {
		return err
	}
	s.cron.Start()
	logs.Info("Scheduler started, cron_expression=" + cronExpr + " (Asia/Bangkok)")
	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) runAll() {
	prompts, err := s.promptRepository.GetAllAcrossCompanies()
	if err != nil {
		logs.Error(err)
		return
	}

	logs.Info(fmt.Sprintf("Scheduler: sweeping %d prompts", len(prompts)))

	ran, failed := 0, 0
	for i, p := range prompts {
		if _, err := s.promptRunService.RunSystem(p.Id, service.TriggerScheduled); err != nil {
			logs.Error(fmt.Errorf("scheduler: run failed for prompt %d: %w", p.Id, err))
			failed++
		} else {
			ran++
		}
		if i < len(prompts)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	logs.Info(fmt.Sprintf("Scheduler: sweep complete, ran=%d failed=%d", ran, failed))
}
