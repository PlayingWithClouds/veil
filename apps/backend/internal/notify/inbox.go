package notify

import (
	"context"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// defaultPageSize is the page size when the caller asks for none.
const defaultPageSize = 50

// Notifications lists a page of the inbox, newest first; unreadOnly leaves
// out the ones already read.
func (s *Service) Notifications(ctx context.Context, unreadOnly bool, limit, offset int) ([]schema.Notification, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	conditions := []string{`"deletedAt" IS NULL`}
	if unreadOnly {
		conditions = append(conditions, `"read" = 0`)
	}
	return db.QueryAs[schema.Notification](ctx, s.database,
		`SELECT * FROM "Notification" WHERE `+strings.Join(conditions, " AND ")+
			` ORDER BY "at" DESC, rowid DESC LIMIT $limit OFFSET $offset`,
		db.Vars{"limit": limit, "offset": offset})
}

// MarkRead marks the notifications with ids as read.
func (s *Service) MarkRead(ctx context.Context, ids []string) error {
	_, err := s.database.Exec(ctx,
		`UPDATE "Notification" SET "read" = 1, "updatedAt" = $now
		 WHERE "read" = 0 AND "id" IN (SELECT value FROM json_each($ids))`,
		db.Vars{"ids": ids, "now": db.Now()})
	return err
}

// MarkAllRead marks the whole inbox as read.
func (s *Service) MarkAllRead(ctx context.Context) error {
	_, err := s.database.Exec(ctx,
		`UPDATE "Notification" SET "read" = 1, "updatedAt" = $now WHERE "read" = 0`,
		db.Vars{"now": db.Now()})
	return err
}

// UnreadCount counts the unread notifications.
func (s *Service) UnreadCount(ctx context.Context) (int, error) {
	return s.database.Int(ctx, `SELECT COUNT(*) FROM "Notification" WHERE "read" = 0 AND "deletedAt" IS NULL`, nil)
}
