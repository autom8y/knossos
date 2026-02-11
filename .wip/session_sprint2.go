package session

import (
	"crypto/rand"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
	"gopkg.in/yaml.v3"
)

// SprintStatus represents sprint lifecycle state.
type SprintStatus string

const (
	// SprintStatusActive indicates an active sprint.
	SprintStatusActive SprintStatus = "ACTIVE"
	// SprintStatusCompleted indicates a completed sprint.
	SprintStatusCompleted SprintStatus = "COMPLETED"
)

// SprintTaskStatus represents task state within a sprint.
type SprintTaskStatus string

const (
	TaskStatusPending    SprintTaskStatus = "pending"
	TaskStatusInProgress SprintTaskStatus = "in_progress"
	TaskStatusDone       SprintTaskStatus = "done"
	TaskStatusSkipped    SprintTaskStatus = "skipped"
)

// Sprint ID format: sprint-YYYYMMDD-HHMMSS-{8-hex}
var sprintIDPattern = regexp.MustCompile(`^sprint-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$`)

// GenerateSprintID generates a new unique sprint ID.
func GenerateSprintID() string {
	now := time.Now()
	hex := make([]byte, 4)
	rand.Read(hex)
	return fmt.Sprintf("sprint-%s-%x",
		now.Format("20060102-150405"),
		hex,
	)
}

// IsValidSprintID checks if an ID matches the sprint ID pattern.
func IsValidSprintID(id string) bool {
	return sprintIDPattern.MatchString(id)
}

// Sprint event types.
const (
	EventSprintCreated   EventType = "SPRINT_CREATED"
	EventSprintStarted   EventType = "SPRINT_STARTED"
	EventSprintCompleted EventType = "SPRINT_COMPLETED"
	EventSprintDeleted   EventType = "SPRINT_DELETED"
)

// EmitSprintCreated emits a SPRINT_CREATED event.
func (e *EventEmitter) EmitSprintCreated(sessionID, sprintID, goal string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSprintCreated,
		Metadata: map[string]interface{}{
			"sprint_id": sprintID,
			"goal":      goal,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitSprintStarted emits a SPRINT_STARTED event.
func (e *EventEmitter) EmitSprintStarted(sessionID, sprintID string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSprintStarted,
		Metadata: map[string]interface{}{
			"sprint_id": sprintID,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitSprintCompleted emits a SPRINT_COMPLETED event.
func (e *EventEmitter) EmitSprintCompleted(sessionID, sprintID string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSprintCompleted,
		Metadata: map[string]interface{}{
			"sprint_id": sprintID,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// EmitSprintDeleted emits a SPRINT_DELETED event.
func (e *EventEmitter) EmitSprintDeleted(sessionID, sprintID string) error {
	event := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     EventSprintDeleted,
		Metadata: map[string]interface{}{
			"sprint_id": sprintID,
		},
	}
	if err := e.Emit(event); err != nil {
		return err
	}
	return e.EmitToAudit(sessionID, event)
}

// SprintTask represents a task within a sprint.
type SprintTask struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Status      string `yaml:"status"`
	Agent       string `yaml:"agent,omitempty"`
}

// SprintContext represents parsed SPRINT_CONTEXT.md.
type SprintContext struct {
	SchemaVersion string       `yaml:"schema_version"`
	SprintID      string       `yaml:"sprint_id"`
	SessionID     string       `yaml:"session_id"`
	Goal          string       `yaml:"goal"`
	Status        SprintStatus `yaml:"status"`
	CreatedAt     time.Time    `yaml:"created_at"`
	CompletedAt   *time.Time   `yaml:"completed_at,omitempty"`
	Tasks         []SprintTask `yaml:"tasks,omitempty"`
	Body          string       `yaml:"-"`
}

type sprintYAML struct {
	SchemaVersion string       `yaml:"schema_version"`
	SprintID      string       `yaml:"sprint_id"`
	SessionID     string       `yaml:"session_id"`
	Goal          string       `yaml:"goal"`
	Status        string       `yaml:"status"`
	CreatedAt     string       `yaml:"created_at"`
	CompletedAt   string       `yaml:"completed_at,omitempty"`
	Tasks         []SprintTask `yaml:"tasks,omitempty"`
}

// NewSprintContext creates a new sprint context with defaults.
func NewSprintContext(sessionID, goal string, taskDescs []string) *SprintContext {
	sprintID := GenerateSprintID()
	now := time.Now().UTC()

	ctx := &SprintContext{
		SchemaVersion: "1.0",
		SprintID:      sprintID,
		SessionID:     sessionID,
		Goal:          goal,
		Status:        SprintStatusActive,
		CreatedAt:     now,
		Body:          defaultSprintBody(goal),
	}

	for i, desc := range taskDescs {
		ctx.Tasks = append(ctx.Tasks, SprintTask{
			ID:          fmt.Sprintf("task-%03d", i+1),
			Description: desc,
			Status:      string(TaskStatusPending),
		})
	}

	return ctx
}

// ParseSprintContext parses SPRINT_CONTEXT.md content.
func ParseSprintContext(content []byte) (*SprintContext, error) {
	str := string(content)

	if !strings.HasPrefix(str, "---\n") && !strings.HasPrefix(str, "---\r\n") {
		return nil, errors.New(errors.CodeSchemaInvalid, "no YAML frontmatter found")
	}

	endIdx := strings.Index(str[4:], "\n---")
	if endIdx == -1 {
		endIdx = strings.Index(str[4:], "\r\n---")
	}
	if endIdx == -1 {
		return nil, errors.New(errors.CodeSchemaInvalid, "unclosed YAML frontmatter")
	}

	yamlContent := str[4 : endIdx+4]
	body := ""
	afterFrontmatter := endIdx + 4 + 4
	if len(str) > afterFrontmatter {
		body = str[afterFrontmatter:]
	}

	var yamlData sprintYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid YAML frontmatter", err)
	}

	ctx := &SprintContext{
		SchemaVersion: yamlData.SchemaVersion,
		SprintID:      yamlData.SprintID,
		SessionID:     yamlData.SessionID,
		Goal:          yamlData.Goal,
		Status:        SprintStatus(yamlData.Status),
		Tasks:         yamlData.Tasks,
		Body:          body,
	}

	if yamlData.CreatedAt != "" {
		t, err := time.Parse(time.RFC3339, yamlData.CreatedAt)
		if err != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid created_at timestamp", err)
		}
		ctx.CreatedAt = t
	}

	if yamlData.CompletedAt != "" {
		t, err := time.Parse(time.RFC3339, yamlData.CompletedAt)
		if err != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid completed_at timestamp", err)
		}
		ctx.CompletedAt = &t
	}

	return ctx, nil
}

// LoadSprintContext loads a sprint context from a file.
func LoadSprintContext(path string) (*SprintContext, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.CodeFileNotFound, "sprint context file not found")
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read sprint context", err)
	}
	return ParseSprintContext(content)
}

// Serialize converts the context back to SPRINT_CONTEXT.md format.
func (c *SprintContext) Serialize() ([]byte, error) {
	yamlData := sprintYAML{
		SchemaVersion: c.SchemaVersion,
		SprintID:      c.SprintID,
		SessionID:     c.SessionID,
		Goal:          c.Goal,
		Status:        string(c.Status),
		CreatedAt:     c.CreatedAt.UTC().Format(time.RFC3339),
		Tasks:         c.Tasks,
	}

	if c.CompletedAt != nil {
		yamlData.CompletedAt = c.CompletedAt.UTC().Format(time.RFC3339)
	}

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to marshal sprint context", err)
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n")
	if c.Body != "" {
		b.WriteString(c.Body)
	}

	return []byte(b.String()), nil
}

// Save writes the context to a file.
func (c *SprintContext) Save(path string) error {
	data, err := c.Serialize()
	if err != nil {
		return err
	}
	if err := fileutil.AtomicWriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to write sprint context", err)
	}
	return nil
}

// MarkTaskComplete marks a task as done.
func (c *SprintContext) MarkTaskComplete(taskID string) error {
	for i := range c.Tasks {
		if c.Tasks[i].ID == taskID {
			if c.Tasks[i].Status == string(TaskStatusDone) {
				return errors.New(errors.CodeLifecycleViolation,
					fmt.Sprintf("task %s is already completed", taskID))
			}
			c.Tasks[i].Status = string(TaskStatusDone)
			return nil
		}
	}
	return errors.New(errors.CodeFileNotFound, fmt.Sprintf("task %s not found", taskID))
}

// AllTasksDone returns true if all tasks are done or skipped.
func (c *SprintContext) AllTasksDone() bool {
	if len(c.Tasks) == 0 {
		return false
	}
	for _, t := range c.Tasks {
		if t.Status != string(TaskStatusDone) && t.Status != string(TaskStatusSkipped) {
			return false
		}
	}
	return true
}

func defaultSprintBody(goal string) string {
	return fmt.Sprintf("\n# Sprint: %s\n\n## Progress\n- Started\n\n## Notes\n\n", goal)
}
