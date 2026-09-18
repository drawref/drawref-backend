package drawref

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultLogLinesPerPage = 150

type DRDatabase struct {
	pool *pgxpool.Pool
}

var TheDb *DRDatabase

func OpenDatabase(connectionUrl string) error {
	pool, err := pgxpool.New(context.Background(), connectionUrl)
	if err != nil {
		return err
	}

	TheDb = &DRDatabase{
		pool,
	}

	return nil
}

func (db *DRDatabase) Close() {
	db.pool.Close()
}

// log lines

func (db *DRDatabase) AddLogLine(logLevel LogLevel, eventType string, message string, data interface{}) error {
	_, err := db.pool.Exec(context.Background(), `
insert into logs (log_level, event_type, message, extra_data)
values ($1, $2, $3, $4)
	`, logLevel, eventType, message, data)

	return err
}

func (db *DRDatabase) GetLogs(page int) (logs []LogLine, err error) {
	startLog := max(page, 0) * DefaultLogLinesPerPage

	rows, err := db.pool.Query(context.Background(), `
SELECT id, ts, log_level, event_type, message, extra_data
FROM logs
ORDER BY ts desc
LIMIT $1
OFFSET $2
`, DefaultLogLinesPerPage, startLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Log Query failed: %v\n", err)
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		var l LogLine

		err = rows.Scan(&l.ID, &l.Time, &l.Level, &l.EventType, &l.Message, &l.Data)
		if err != nil {
			return logs, err
		}
		logs = append(logs, l)
	}

	return logs, err
}
