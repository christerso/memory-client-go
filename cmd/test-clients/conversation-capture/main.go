package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Command-line flags
var (
	serverURL     = flag.String("server", "http://localhost:10010", "Memory client server URL")
	mode          = flag.String("mode", "send", "Mode: send, set-tag, get-tag, set-mode, get-mode")
	messageRole   = flag.String("role", "user", "Message role (user or assistant)")
	content       = flag.String("content", "", "Message content")
	tag           = flag.String("tag", "", "Conversation tag")
	taggingMode   = flag.String("tagging-mode", "automatic", "Tagging mode (automatic or manual)")
	watchStdin    = flag.Bool("watch", false, "Watch stdin for messages (format: ROLE: CONTENT)")
	logFile       = flag.String("log", "", "Log file path")
	showVersion   = flag.Bool("version", false, "Show version information")
	ClientVersion = "1.3.0"
)

const ()

func init() {
	// Define command-line flags
	flag.StringVar(serverURL, "server", "http://localhost:10010", "Memory client server URL")
	flag.StringVar(mode, "mode", "send", "Mode: send, set-tag, get-tag, set-mode, get-mode")
	flag.StringVar(messageRole, "role", "user", "Message role (user or assistant)")
	flag.StringVar(content, "content", "", "Message content")
	flag.StringVar(tag, "tag", "", "Conversation tag")
	flag.StringVar(taggingMode, "tagging-mode", "automatic", "Tagging mode (automatic or manual)")
	flag.BoolVar(watchStdin, "watch", false, "Watch stdin for messages (format: ROLE: CONTENT)")
	flag.StringVar(logFile, "log", "", "Log file path")
	flag.BoolVar(showVersion, "version", false, "Show version information")
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("Conversation Capture Client v%s\n", ClientVersion)
		return
	}

	// Set up logging
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Validate mode
	switch *mode {
	case "send":
		if *content == "" {
			log.Fatal("Error: content is required for send mode")
		}
		if *messageRole != "user" && *messageRole != "assistant" {
			log.Fatal("Error: role must be 'user' or 'assistant'")
		}
		sendMessage(*serverURL, *messageRole, *content)
	case "set-tag":
		if *tag == "" {
			log.Fatal("Error: tag is required for set-tag mode")
		}
		setConversationTag(*serverURL, *tag)
	case "get-tag":
		getCurrentTag(*serverURL)
	case "set-mode":
		if *taggingMode != "automatic" && *taggingMode != "manual" {
			log.Fatal("Error: tagging-mode must be 'automatic' or 'manual'")
		}
		setTaggingMode(*serverURL, *taggingMode)
	case "get-mode":
		getTaggingMode(*serverURL)
	case "watch":
		watchStdinForMessages()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Conversation Capture Client")
	fmt.Println("Usage:")
	fmt.Println("  Send message:      conversation-capture-client -mode=send -role=<role> -content=<content>")
	fmt.Println("  Set tag:           conversation-capture-client -mode=set-tag -tag=<tag>")
	fmt.Println("  Get current tag:   conversation-capture-client -mode=get-tag")
	fmt.Println("  Set tagging mode:   conversation-capture-client -mode=set-mode -tagging-mode=<tagging-mode>")
	fmt.Println("  Get tagging mode:   conversation-capture-client -mode=get-mode")
	fmt.Println("  Watch stdin:       conversation-capture-client -watch")
	fmt.Println("")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  conversation-capture-client -mode=send -role=user -content=\"How do I create a Go struct?\"")
	fmt.Println("  conversation-capture-client -mode=set-tag -tag=golang-tutorial")
	fmt.Println("  echo \"user: Hello there\" | conversation-capture-client -watch")
}

func watchStdinForMessages() {
	fmt.Println("Watching stdin for messages. Format: ROLE: CONTENT")
	fmt.Println("Press Ctrl+C to stop")

	scanner := NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			log.Printf("Invalid format: %s", line)
			continue
		}

		role := strings.TrimSpace(parts[0])
		content := strings.TrimSpace(parts[1])

		if role == "" || content == "" {
			log.Printf("Invalid message: role or content is empty")
			continue
		}

		sendMessage(*serverURL, role, content)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading stdin: %v", err)
	}
}

func sendMessage(serverURL, role, content string) {
	// Create the request payload
	payload := map[string]interface{}{
		"role":    role,
		"content": content,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Send the request to the memory client server
	resp, err := http.Post(serverURL+"/api/message", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request to memory client server: %v", err)
	}
	defer resp.Body.Close()

	// Read and process the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error response from server: %s", string(body))
	}

	// Parse the response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Warning: Could not parse response: %v", err)
	}

	log.Printf("Message sent successfully: %s: %s", role, truncateString(content, 50))
	fmt.Printf("Message sent: %s: %s\n", role, truncateString(content, 50))
}

func setConversationTag(serverURL, tag string) {
	// Create the request payload
	payload := map[string]interface{}{
		"tag": tag,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Send the request to the memory client server
	resp, err := http.Post(serverURL+"/api/set-conversation-tag", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request to memory client server: %v", err)
	}
	defer resp.Body.Close()

	// Read and process the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error response from server: %s", string(body))
	}

	log.Printf("Conversation tag set to: %s", tag)
	fmt.Printf("Conversation tag set to: %s\n", tag)
}

func getCurrentTag(serverURL string) {
	// Send the request to the memory client server
	resp, err := http.Get(serverURL + "/api/get-conversation-tag")
	if err != nil {
		log.Fatalf("Error sending request to memory client server: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	var result struct {
		Tag string `json:"tag"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	if result.Tag == "" {
		log.Println("No conversation tag is currently set")
		fmt.Println("No conversation tag is currently set")
	} else {
		log.Printf("Current conversation tag: %s", result.Tag)
		fmt.Printf("Current conversation tag: %s\n", result.Tag)
	}
}

func setTaggingMode(serverURL, mode string) {
	// Create request body
	requestBody, err := json.Marshal(map[string]string{
		"mode": mode,
	})
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Send request
	resp, err := http.Post(
		fmt.Sprintf("%s/api/set-tagging-mode", serverURL),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: %s", body)
	}

	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	// Print response
	fmt.Println(response["message"])
}

func getTaggingMode(serverURL string) {
	// Send request
	resp, err := http.Get(fmt.Sprintf("%s/api/get-tagging-mode", serverURL))
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: %s", body)
	}

	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	// Print response
	fmt.Printf("Current tagging mode: %s\n", response["mode"])
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Scanner is a custom scanner that handles large lines
type Scanner struct {
	reader    io.Reader
	buf       []byte
	start     int
	end       int
	err       error
	maxBuffer int
	token     []byte
}

// NewScanner creates a new scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		reader:    r,
		buf:       make([]byte, 4096),
		maxBuffer: 1024 * 1024, // 1MB max buffer
	}
}

// Scan advances the scanner to the next token
func (s *Scanner) Scan() bool {
	s.token = nil
	s.err = nil

	for {
		// If we have data in the buffer, try to find a newline
		if s.start < s.end {
			i := s.start
			for i < s.end {
				if s.buf[i] == '\n' {
					s.token = s.buf[s.start:i]
					s.start = i + 1
					return true
				}
				i++
			}

			// No newline found, check if buffer is full
			if s.end == len(s.buf) {
				// If start is at the beginning, we need to grow the buffer
				if s.start == 0 {
					// Check if we've reached the maximum buffer size
					if len(s.buf) >= s.maxBuffer {
						s.err = fmt.Errorf("line too long (max %d bytes)", s.maxBuffer)
						return false
					}

					// Grow the buffer
					newBuf := make([]byte, len(s.buf)*2)
					copy(newBuf, s.buf)
					s.buf = newBuf
				} else {
					// Shift data to the beginning of the buffer
					copy(s.buf, s.buf[s.start:s.end])
					s.end -= s.start
					s.start = 0
				}
			}
		}

		// Read more data
		n, err := s.reader.Read(s.buf[s.end:])
		if n > 0 {
			s.end += n
			continue
		}

		if err == io.EOF {
			// End of file, return any remaining data
			if s.start < s.end {
				s.token = s.buf[s.start:s.end]
				s.start = s.end
				return true
			}
			return false
		}

		s.err = err
		return false
	}
}

// Text returns the current token
func (s *Scanner) Text() string {
	return string(s.token)
}

// Err returns the current error
func (s *Scanner) Err() error {
	return s.err
}
