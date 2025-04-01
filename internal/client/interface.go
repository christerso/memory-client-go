package client

import (
	"context"
	"time"

	"github.com/christerso/memory-client-go/internal/models"
)

// MemoryClientInterface defines the interface for memory client operations
type MemoryClientInterface interface {
	// General methods
	Close() error
	
	// Message operations
	AddMessage(ctx context.Context, message *models.Message) error
	GetConversationHistory(ctx context.Context, limit int, filter *models.HistoryFilter) ([]models.Message, error)
	SearchMessages(ctx context.Context, query string, limit int) ([]models.Message, error)
	DeleteMessage(ctx context.Context, id string) error
	DeleteAllMessages(ctx context.Context) error
	DeleteMessagesForCurrentDay(ctx context.Context) (int, error)
	DeleteMessagesForCurrentWeek(ctx context.Context) (int, error)
	DeleteMessagesForCurrentMonth(ctx context.Context) (int, error)
	DeleteMessagesByTimeRange(ctx context.Context, from, to time.Time) (int, error)
	TagMessages(ctx context.Context, ids []string, tag string) error
	GetMessagesByTag(ctx context.Context, tag string, limit int) ([]models.Message, error)
	IndexMessages(ctx context.Context) error
	
	// Project file operations
	IndexProjectFiles(ctx context.Context, projectPath, tag string) (int, error)
	UpdateProjectFiles(ctx context.Context, projectPath string) (int, int, error)
	SearchProjectFiles(ctx context.Context, query string, limit int) ([]models.ProjectFile, error)
	ListProjectFiles(ctx context.Context, limit int) ([]models.ProjectFile, error)
	ListProjectFilesByTag(ctx context.Context, tag string, limit int) ([]models.ProjectFile, error)
	DeleteProjectFile(ctx context.Context, id string) error
	DeleteAllProjectFiles(ctx context.Context) error
	DeleteProjectFilesByTag(ctx context.Context, tag string) error
	
	// Memory clearing operations
	ClearAllMemories(ctx context.Context) error
	ClearMessages(ctx context.Context) error
	ClearProjectFiles(ctx context.Context) error
	
	// Utility operations
	SummarizeAndTagMessages(ctx context.Context, timeRange models.TimeRange, tag string) (string, error)
	GetMemoryStats(ctx context.Context) (*models.MemoryStats, error)
	PurgeQdrant(ctx context.Context) error
}
