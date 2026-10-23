package recommendeval

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// copyDatabase writes a consistent snapshot of the database at sourcePath to
// destinationPath without modifying the source: a read-only connection and
// VACUUM INTO, which also folds in whatever sits in the write-ahead log.
func copyDatabase(ctx context.Context, sourcePath, destinationPath string) error {
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("database %s: %w", sourcePath, err)
	}
	source, err := sql.Open("sqlite", "file:"+sourcePath+"?mode=ro&_pragma=busy_timeout(10000)")
	if err != nil {
		return fmt.Errorf("open %s read-only: %w", sourcePath, err)
	}
	defer source.Close()
	quotedDestination := "'" + strings.ReplaceAll(destinationPath, "'", "''") + "'"
	if _, err := source.ExecContext(ctx, "VACUUM INTO "+quotedDestination); err != nil {
		return fmt.Errorf("copy %s: %w", sourcePath, err)
	}
	return nil
}
