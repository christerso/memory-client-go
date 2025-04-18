package models

import (
	"fmt"
	"time"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleProject   Role = "project" // Special role for project files
)

// Message represents a conversation message
type Message struct {
	ID        string            `json:"id"`
	Role      Role              `json:"role"`
	Content   string            `json:"content"`
	Embedding []float32         `json:"embedding"`
	Tags      []string          `json:"tags,omitempty"`
	Summary   string            `json:"summary,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Score     float64           `json:"score,omitempty"` // For search results
}

// ProjectFile represents a file in a project
type ProjectFile struct {
	ID        string    `json:"id"`              // Unique identifier
	Path      string    `json:"path"`            // Relative path to the file
	Content   string    `json:"content"`         // File content
	Language  string    `json:"language"`        // Programming language or file type
	Vector    []float32 `json:"-"`               // Vector embedding
	ModTime   int64     `json:"mod_time"`        // Last modification time (Unix timestamp)
	Tag       string    `json:"tag,omitempty"`   // Optional tag for categorization
	Timestamp time.Time `json:"timestamp"`       // Time when the file was indexed
	Score     float64   `json:"score,omitempty"` // For search results
}

// HistoryFilter represents a filter for conversation history
type HistoryFilter struct {
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Role      Role      `json:"role,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
}

// TimeRange represents a time range for operations
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	TotalVectors     int            `json:"total_vectors"`
	MessageCount     map[string]int `json:"message_count"`
	ProjectFileCount int            `json:"project_file_count"`
}

// MediaExtensions is a list of file extensions to exclude from project indexing
var MediaExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
	".svg": true, ".ico": true, ".webp": true, ".tiff": true, ".tif": true,
	".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".aac": true,
	".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true,
	".webm": true, ".mkv": true, ".m4v": true, ".3gp": true,
	".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".bin": true,
	".dat": true, ".db": true, ".sqlite": true, ".mdb": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true,
}

// BinaryExtensions is a list of binary file extensions to exclude
var BinaryExtensions = map[string]bool{
	".o": true, ".a": true, ".lib": true, ".obj": true, ".class": true,
	".pyc": true, ".pyo": true, ".pyd": true,
}

// LanguageMap maps file extensions to language names
var LanguageMap = map[string]string{
	".go":     "Go",
	".py":     "Python",
	".js":     "JavaScript",
	".ts":     "TypeScript",
	".jsx":    "JavaScript (React)",
	".tsx":    "TypeScript (React)",
	".html":   "HTML",
	".css":    "CSS",
	".scss":   "SCSS",
	".less":   "LESS",
	".json":   "JSON",
	".xml":    "XML",
	".yaml":   "YAML",
	".yml":    "YAML",
	".md":     "Markdown",
	".txt":    "Text",
	".sh":     "Shell",
	".bat":    "Batch",
	".ps1":    "PowerShell",
	".c":      "C",
	".cpp":    "C++",
	".h":      "C/C++ Header",
	".hpp":    "C++ Header",
	".cs":     "C#",
	".java":   "Java",
	".rb":     "Ruby",
	".php":    "PHP",
	".swift":  "Swift",
	".kt":     "Kotlin",
	".rs":     "Rust",
	".sql":    "SQL",
	".r":      "R",
	".dart":   "Dart",
	".lua":    "Lua",
	".scala":  "Scala",
	".pl":     "Perl",
	".groovy": "Groovy",
	".elm":    "Elm",
	".ex":     "Elixir",
	".exs":    "Elixir",
	".erl":    "Erlang",
	".hs":     "Haskell",
	".fs":     "F#",
	".fsx":    "F#",
	".clj":    "Clojure",
	".toml":   "TOML",
	".ini":    "INI",
	".cfg":    "Configuration",
	".conf":   "Configuration",
}

// NewMessage creates a new message with the given role and content
func NewMessage(role Role, content string) *Message {
	return &Message{
		ID:        generateUUID(),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// generateUUID generates a unique ID in UUID v4 format
func generateUUID() string {
	// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	// Where x is any hexadecimal digit and y is one of 8, 9, A, or B

	// Generate 16 random bytes
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> uint(i*8) & 0xff)
	}

	// Set version to 4 (random UUID)
	b[6] = (b[6] & 0x0f) | 0x40

	// Set variant to RFC4122
	b[8] = (b[8] & 0x3f) | 0x80

	// Format as UUID string
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
