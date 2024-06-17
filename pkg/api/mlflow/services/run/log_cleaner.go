package run

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/config"
)

// LogsCleanerProvider provides an interface to work with LogCleaner.
type LogsCleanerProvider interface {
	// Run runs manager background jobs.
	Run(ctx context.Context)
}

// LogCleaner represents Run Logs cleaner.
type LogCleaner struct {
	ctx           context.Context
	config        *config.Config
	logRepository repositories.LogRepositoryProvider
}

// NewLogCleaner creates a new instance of LogCleaner.
func NewLogCleaner(
	ctx context.Context,
	config *config.Config,
	logRepository repositories.LogRepositoryProvider,
) *LogCleaner {
	return &LogCleaner{
		ctx:           ctx,
		config:        config,
		logRepository: logRepository,
	}
}

// Run run logs cleaner background jobs.
func (m LogCleaner) Run() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-m.ctx.Done():
				log.Debug("run logs cleaner finished. exiting.")
				return
			case <-ticker.C:
				if m.config.RunLogOutputMax != 0 {
					numberOfDeleted, err := m.logRepository.CleanExpired(m.ctx, m.config.RunLogOutputRetain)
					if err != nil {
						log.Errorf("error cleaning expired run logs: %+v", err)
					} else {
						log.Debugf("%d expired run logs were successfully cleaned", numberOfDeleted)
					}
				}
			default:
				time.Sleep(5 * time.Minute)
			}
		}
	}()
}
