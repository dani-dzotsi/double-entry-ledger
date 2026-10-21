package jobs

import "time"

type JobType string

const POST_TRANSFER JobType = "POST_TRANSFER"

type JobStatus string

const (
	QUEUED  JobStatus = "QUEUED"
	RUNNING JobStatus = "RUNNING"
	DONE    JobStatus = "DONE"
	FAILED  JobStatus = "FAILED"
)

// JobPayload is stored as PostgreSQL JSONB.
type JobPayload map[string]any

type Job struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;type:bigint"`
	Type        JobType    `gorm:"type:text;not null"`
	Payload     JobPayload `gorm:"type:jsonb;not null"`
	Status      JobStatus  `gorm:"type:text;not null;check:chk_jobs_status,status IN ('QUEUED','RUNNING','DONE','FAILED')"`
	Attempts    int        `gorm:"type:integer;not null;default:0"`
	MaxAttempts int        `gorm:"type:integer;not null;default:5"`
	RunAt       time.Time  `gorm:"type:timestamptz;not null;index:idx_jobs_status_run_at,priority:2"`
	LockedAt    *time.Time `gorm:"type:timestamptz"`
	LastError   *string    `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;not null;index:idx_jobs_status_run_at,priority:1"`
}
