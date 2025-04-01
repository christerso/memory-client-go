package main

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleProject   Role = "project" // Special role for project files
)

// Message represents a conversation message
type Message struct {
	Role    Role        `json:"role"`
	Content string      `json:"content"`
	Vector  []float32   `json:"-"`
	Tags    []string    `json:"tags,omitempty"`
	Summary string      `json:"summary,omitempty"`
}

// ProjectFile represents a file in a project
type ProjectFile struct {
	Path     string    `json:"path"`     // Relative path to the file
	Content  string    `json:"content"`  // File content
	Language string    `json:"language"` // Programming language or file type
	Vector   []float32 `json:"-"`        // Vector embedding
	ModTime  int64     `json:"mod_time"` // Last modification time (Unix timestamp)
	Tag      string    `json:"tag,omitempty"` // Optional tag for categorization
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

func NewMessage(role Role, content string) *Message {
	return &Message{
		Role:    role,
		Content: content,
	}
}
