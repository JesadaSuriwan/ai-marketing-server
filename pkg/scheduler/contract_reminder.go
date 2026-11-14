package scheduler

import (
	"fmt"
	"time"

	"github.com/ai-marketing/ai-marketing-server/logs"
	notificationRepository "github.com/ai-marketing/ai-marketing-server/pkg/notification/repository"
)

// NotificationTypeContract is the notification "type" used for
// contract-expiry reminders — the frontend uses it to decide both the icon
// shown in the Notification Center and the deep link ("View contract")
// surfaced in the push reminder.
const NotificationTypeContract = "contract"

const contractReminderWindowDays = 7

// parseContractEndDate accepts both a bare "2006-01-02" date and the
// RFC3339 timestamp (e.g. "2026-11-10T00:00:00Z") that lib/pq actually
// returns when scanning a Postgres `date` column into a Go string.
func parseContractEndDate(raw string) (time.Time, error) {
	if end, err := time.Parse("2006-01-02", raw); err == nil {
		return end, nil
	}
	return time.Parse(time.RFC3339, raw)
}

// checkContractReminders runs once a day (same cron cadence as the prompt
// sweep) and creates one "contract" notification per company whose
// contract_end_date falls within the next contractReminderWindowDays days.
// It's deduped by the notification's exact description text, which embeds
// the end date — so renewing (which changes contract_end_date) is what
// makes a new reminder eligible to fire again, rather than re-notifying
// every day for the same expiry.
func (s *Scheduler) checkContractReminders() {
	companies, err := s.companyRepository.GetAllWithContractEndDate()
	if err != nil {
		logs.Error(fmt.Errorf("contract reminder sweep: %w", err))
		return
	}

	created := 0
	for _, c := range companies {
		if c.ContractEndDate == nil {
			continue
		}
		end, err := parseContractEndDate(*c.ContractEndDate)
		if err != nil {
			logs.Error(fmt.Errorf("contract reminder: company %d has unparseable contract_end_date %q: %w", c.Id, *c.ContractEndDate, err))
			continue
		}

		daysLeft := int(time.Until(end).Hours() / 24)
		if daysLeft < 0 || daysLeft > contractReminderWindowDays {
			continue
		}

		description := fmt.Sprintf(
			"%s's contract ends on %s. Renew it in Settings to keep prompt runs running without interruption.",
			c.Name, end.Format("January 2, 2006"),
		)

		exists, err := s.notificationRepository.ExistsByCompanyTypeAndDescription(c.Id, NotificationTypeContract, description)
		if err != nil {
			logs.Error(fmt.Errorf("contract reminder: company %d: %w", c.Id, err))
			continue
		}
		if exists {
			continue
		}

		tx, err := s.notificationRepository.NewTransaction()
		if err != nil {
			logs.Error(fmt.Errorf("contract reminder: company %d: %w", c.Id, err))
			continue
		}
		_, err = s.notificationRepository.Create(tx, notificationRepository.Notification{
			UserId: c.UserId, CompanyId: c.Id, Type: NotificationTypeContract,
			Title: "Contract ending soon", Description: description,
		})
		if err != nil {
			tx.Rollback()
			logs.Error(fmt.Errorf("contract reminder: company %d: %w", c.Id, err))
			continue
		}
		if err := tx.Commit(); err != nil {
			logs.Error(fmt.Errorf("contract reminder: company %d: %w", c.Id, err))
			continue
		}
		created++
	}

	if created > 0 {
		logs.Info(fmt.Sprintf("Contract reminder sweep: created %d notification(s)", created))
	}
}
