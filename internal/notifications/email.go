package notifications

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"bereaucat/internal/mailer"
	"bereaucat/internal/store"
)

const emailTick = time.Minute

// activitySlack widens the notification's lifetime when matching activity rows,
// since activity is logged slightly before its notification row is written.
const activitySlack = 5 * time.Second

// RunEmailDigest emails each notification once its coalescing window closes, until ctx is cancelled.
func (s *Service) RunEmailDigest(ctx context.Context) {
	t := time.NewTicker(emailTick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.sendEmailDigest(ctx)
		}
	}
}

func (s *Service) sendEmailDigest(ctx context.Context) {
	cfg, err := mailer.Load(ctx, s.store)
	if err != nil {
		log.Printf("notifications: failed to load smtp settings: %v", err)
		return
	}

	cutoff := pgtype.Timestamptz{Time: time.Now().Add(-s.window), Valid: true}

	if cfg.Enabled {
		rows, err := s.store.ClaimEmailNotifications(ctx, cutoff)
		if err != nil {
			log.Printf("notifications: failed to claim email notifications: %v", err)
			return
		}
		for _, row := range rows {
			msg, err := notificationEmail(cfg.AppURL, row, s.emailEvents(ctx, row))
			if err == nil {
				err = mailer.Send(ctx, cfg, msg)
			}
			if err != nil {
				log.Printf("notifications: failed to email notification %s: %v", row.ID, err)
			}
		}
	}

	if err := s.store.SkipEmailNotifications(ctx, cutoff); err != nil {
		log.Printf("notifications: failed to skip email notifications: %v", err)
	}
}

func (s *Service) emailEvents(ctx context.Context, row store.ClaimEmailNotificationsRow) []emailEvent {
	activity, err := s.store.ListEmailActivity(ctx, store.ListEmailActivityParams{
		TaskID:      row.TaskID,
		RecipientID: row.RecipientID,
		Since:       pgtype.Timestamptz{Time: row.CreatedAt.Time.Add(-activitySlack), Valid: true},
		Until:       pgtype.Timestamptz{Time: row.UpdatedAt.Time.Add(activitySlack), Valid: true},
	})
	if err != nil {
		log.Printf("notifications: failed to load activity for notification %s: %v", row.ID, err)
	}

	events := make([]emailEvent, 0, len(activity))
	for _, a := range activity {
		events = append(events, eventFromActivity(a))
	}
	if len(events) == 0 {
		events = append(events, newEvent(row.ActorFirstName, row.ActorLastName, activityLabel(row.ActivityType)))
	}
	return events
}
