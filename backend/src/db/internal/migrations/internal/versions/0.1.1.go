package versions

import (
	"luna-backend/db/internal/migrations/internal/registry"
	migrationTypes "luna-backend/db/internal/migrations/types"
	"luna-backend/errors"
	"luna-backend/types"
)

func init() {
	registry.RegisterMigration(types.Ver(0, 1, 1), func(q *migrationTypes.MigrationQueries) *errors.ErrorTrace {
		// Add the location column to event_overrides for the new event location feature
		_, err := q.Tx.Exec(
			q.Context,
			`
			ALTER TABLE event_overrides ADD COLUMN IF NOT EXISTS location TEXT;
			`,
		)

		if err != nil {
			return errors.New().
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not add location column to event_overrides")
		}

		return nil
	})
}
