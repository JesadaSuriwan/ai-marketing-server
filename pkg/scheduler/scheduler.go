package scheduler

import (
	"fmt"
	"time"

	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptrun/service"
	"github.com/robfig/cron/v3"
)

// Scheduler runs every tracked prompt (across every company) on a cron
// cadence, reusing the same run + citation-extraction pipeline as the
// manual "Run Prompt" button and the "run all" endpoint.
type Scheduler struct {
	cron             *cron.Cron
	promptRunService service.PromptRunService
	promptRepository repository.PromptRepository
}

func New(promptRunService service.PromptRunService, promptRepository repository.PromptRepository) *Scheduler {
	return &Scheduler{
		cron:             cron.New(),
		promptRunService: promptRunService,
		promptRepository: promptRepository,
	}
}

// Start schedules the sweep on cronExpr (standard 5-field cron syntax) and
// begins running it in the background.
func (s *Scheduler) Start(cronExpr string) error {
	if _, err := s.cron.AddFunc(cronExpr, s.runAll); err != nil {
		return err
	}
	s.cron.Start()
	logs.Info("Scheduler started, cron_expression=" + cronExpr)
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
		if _, err := s.promptRunService.RunSystem(p.Id); err != nil {
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
