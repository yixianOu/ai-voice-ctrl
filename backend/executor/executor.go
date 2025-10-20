package executor

// Executor is the interface that wraps the basic Execute method.
type Executor interface {
	/* P0 */
	// Find files, open folders, and open files, read documents, write notes
	FindFiles(query string) error
	OpenFolder(path string) error
	OpenFile(path string) error
	Read(path string) (string, error)
	Write(content string) error

	// Open, close, and switch applications
	Open(app string) error
	Close(app string) error
	Switch(app string) error

	/* P1 */
	// Play, pause, switch music/video
	Play() error
	Pause() error
	Next() error
	Previous() error

	/* P2 */
	// Web browsing and information search
	SearchInfo(query string) error

	// Dictation, writing emails, writing code snippets
	Dictate(text string) error
	WriteEmail(recipient, subject, body string) error
	WriteCodeSnippet(language, description string) error
}
