// Package notifications implements per-user, persisted in-app notifications with
// read/unread state and write-time coalescing. When several activities happen on
// the same task within a short window, they collapse into a single notification
// row so users are not spammed (max 1 notification per task per window).
package notifications

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"bereaucat/internal/store"
)

// DefaultWindow is the coalescing window: activity on the same task within this
// duration is merged into the recipient's existing notification.
const DefaultWindow = 15 * time.Minute

// Service fans out task activity to per-user notification rows.
type Service struct {
	store  store.Querier
	pool   *pgxpool.Pool
	window time.Duration
}

// NewService creates a notifications service with the default 15-minute window.
func NewService(s store.Querier, pool *pgxpool.Pool) *Service {
	return &Service{store: s, pool: pool, window: DefaultWindow}
}

// EnqueueForActivity creates (or coalesces into) a notification for every
// participant of the task except the actor. It is best-effort: failures are
// logged and never propagated, so notification delivery can never break the
// underlying request.
func (s *Service) EnqueueForActivity(ctx context.Context, taskID uuid.UUID, activityType string, actorID uuid.UUID, commentID *uuid.UUID) {
	participants, err := s.store.ListTaskParticipants(ctx, taskID)
	if err != nil {
		log.Printf("notifications: failed to list participants for task %s: %v", taskID, err)
		return
	}

	cutoff := pgtype.Timestamptz{Time: time.Now().Add(-s.window), Valid: true}

	commentRef := pgtype.UUID{Valid: false}
	if commentID != nil {
		commentRef = pgtype.UUID{Bytes: *commentID, Valid: true}
	}

	for _, recipientID := range participants {
		if recipientID == actorID {
			continue
		}
		if err := s.upsertNotification(ctx, recipientID, taskID, activityType, actorID, commentRef, cutoff); err != nil {
			log.Printf("notifications: failed to notify user %s on task %s: %v", recipientID, taskID, err)
		}
	}
}

// upsertNotification coalesces into the recipient's open notification for the task, or creates one.
func (s *Service) upsertNotification(ctx context.Context, recipientID, taskID uuid.UUID, activityType string, actorID uuid.UUID, commentRef pgtype.UUID, cutoff pgtype.Timestamptz) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := store.New(tx)

	if err := q.LockNotificationSlot(ctx, store.LockNotificationSlotParams{RecipientID: recipientID, TaskID: taskID}); err != nil {
		return err
	}

	open, err := q.GetOpenNotification(ctx, store.GetOpenNotificationParams{
		RecipientID: recipientID,
		TaskID:      taskID,
		Cutoff:      cutoff,
	})
	switch {
	case err == nil:
		err = q.CoalesceNotification(ctx, store.CoalesceNotificationParams{
			ID:           open.ID,
			ActivityType: activityType,
			ActorID:      actorID,
			CommentID:    commentRef,
		})
	case errors.Is(err, pgx.ErrNoRows):
		_, err = q.CreateNotification(ctx, store.CreateNotificationParams{
			RecipientID:  recipientID,
			TaskID:       taskID,
			ActivityType: activityType,
			ActorID:      actorID,
			CommentID:    commentRef,
		})
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
