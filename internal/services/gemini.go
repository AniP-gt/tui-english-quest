package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings" // Added strings import

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient holds the Generative AI client
type GeminiClient struct {
	client *genai.GenerativeModel
}

// NewGeminiClient initializes and returns a new GeminiClient
func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	// Create a new client with the API key
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Select the model (e.g., "gemini-pro")
	model := client.GenerativeModel("gemini-2.5-flash")

	return &GeminiClient{client: model}, nil
}

const (
	ModeVocab     = "vocab"
	ModeGrammar   = "grammar"
	ModeTavern    = "tavern"
	ModeSpelling  = "spelling"
	ModeListening = "listening"
)

// QuestionPayload contains fetched questions for a mode.
type QuestionPayload struct {
	Mode    string
	Content []byte
}

// FetchQuestions fetches five questions for the given mode using the Gemini API.
func FetchQuestions(ctx context.Context, mode string) (QuestionPayload, error) {
	gc, err := NewGeminiClient(ctx)
	if err != nil {
		return QuestionPayload{}, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	// The genai.GenerativeModel does not have a Close() method, so we don't defer it here.

	prompt := fmt.Sprintf("Generate 5 *new and diverse* %s questions in JSON format. The JSON should strictly adhere to the following structure for %s mode:\n\n", mode, mode)

	switch mode {
	case ModeVocab:
		prompt += `
{
  "questions": [
    {
      "enemy_name": "string",
      "word": "string",
      "options": ["string", "string", "string", "string"],
      "answer_index": "integer (0-3)",
      "explanation": "string"
    }
  ]
}`
	case ModeGrammar:
		prompt += `
{
  "traps": [
    {
      "trap_name": "string",
      "question": "string",
      "options": ["string", "string", "string", "string"],
      "answer_index": "integer (0-3)",
      "explanation": "string"
    }
  ]
}`
	case ModeTavern:
		prompt += `
{
  "npc_name": "string",
  "npc_opening": "string",
  "evaluation_rubric": ["string", "string", "string"],
  "turns": [
    {
      "npc_reply": "string"
    }
  ]
}`
	case ModeSpelling:
		prompt += `
{
  "prompts": [
    {
      "ja_hint": "string",
      "correct_spelling": "string",
      "explanation": "string"
    }
  ]
}`
	case ModeListening:
		prompt += `
{
  "audio": [
    {
      "prompt": "string",
      "options": ["string", "string", "string", "string"],
      "answer_index": "integer (0-3)",
      "transcript": "string"
    }
  ]
}`
	default:
		return QuestionPayload{}, fmt.Errorf("unknown mode: %s", mode)
	}

	resp, err := gc.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return QuestionPayload{}, fmt.Errorf("failed to generate content from Gemini API: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return QuestionPayload{}, errors.New("no content found in Gemini API response")
	}

	var contentBytes []byte
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			contentBytes = append(contentBytes, []byte(text)...)
		}
	}

	// The Gemini API might return markdown, so we need to extract the JSON block.
	// This is a simple approach; a more robust solution might use a JSON parser that
	// can handle surrounding text.
	jsonString := string(contentBytes)
	jsonStart := strings.Index(jsonString, "{")
	jsonEnd := strings.LastIndex(jsonString, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		return QuestionPayload{}, fmt.Errorf("could not extract JSON from Gemini API response: %s", jsonString)
	}

	extractedJSON := []byte(jsonString[jsonStart : jsonEnd+1])

	return QuestionPayload{Mode: mode, Content: extractedJSON}, nil
}

// FetchAndValidate obtains a payload then validates schema/count.
func FetchAndValidate(ctx context.Context, mode string) (QuestionPayload, error) {
	payload, err := FetchQuestions(ctx, mode)
	if err != nil {
		return payload, fmt.Errorf("failed to fetch questions for mode %s: %w", mode, err)
	}
	if err := ValidatePayload(payload); err != nil {
		return payload, fmt.Errorf("validation failed for mode %s: %w", mode, err)
	}
	return payload, nil
}

// ValidatePayload validates JSON shape and count (5 items expected).
func ValidatePayload(payload QuestionPayload) error {
	switch payload.Mode {
	case ModeVocab:
		return validateVocab(payload.Content)
	case ModeGrammar:
		return validateGrammar(payload.Content)
	case ModeTavern:
		return validateTavern(payload.Content)
	case ModeSpelling:
		return validateSpelling(payload.Content)
	case ModeListening:
		return validateListening(payload.Content)
	default:
		return fmt.Errorf("unknown mode: %s", payload.Mode)
	}
}

// --- mode validators ---

type VocabQuestion struct {
	EnemyName   string   `json:"enemy_name"`
	Word        string   `json:"word"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Explanation string   `json:"explanation"`
}

type VocabEnvelope struct {
	Questions []VocabQuestion `json:"questions"`
}

func validateVocab(raw []byte) error {
	var env VocabEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("invalid vocab JSON: %w", err)
	}
	if len(env.Questions) != 5 {
		return fmt.Errorf("vocab questions must be 5, got %d", len(env.Questions))
	}
	for i, q := range env.Questions {
		if err := validateOptions(q.Options, q.AnswerIndex); err != nil {
			return fmt.Errorf("vocab question %d: %w", i, err)
		}
	}
	return nil
}

type GrammarTrap struct {
	TrapName    string   `json:"trap_name"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Explanation string   `json:"explanation"`
}

type GrammarEnvelope struct {
	Traps []GrammarTrap `json:"traps"`
}

func validateGrammar(raw []byte) error {
	var env GrammarEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("invalid grammar JSON: %w", err)
	}
	if len(env.Traps) != 5 {
		return fmt.Errorf("grammar traps must be 5, got %d", len(env.Traps))
	}
	for i, t := range env.Traps {
		if err := validateOptions(t.Options, t.AnswerIndex); err != nil {
			return fmt.Errorf("grammar trap %d: %w", i, err)
		}
	}
	return nil
}

type TavernTurn struct {
	NPCReply string `json:"npc_reply"`
}

type TavernEnvelope struct {
	NPCName          string       `json:"npc_name"`
	NPCOpening       string       `json:"npc_opening"`
	EvaluationRubric []string     `json:"evaluation_rubric"`
	Turns            []TavernTurn `json:"turns"`
}

func validateTavern(raw []byte) error {
	var env TavernEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("invalid tavern JSON: %w", err)
	}
	if len(env.Turns) != 5 {
		return fmt.Errorf("tavern turns must be 5, got %d", len(env.Turns))
	}
	if len(env.EvaluationRubric) < 3 {
		return errors.New("tavern evaluation_rubric must have at least 3 levels")
	}
	return nil
}

type SpellingPrompt struct {
	JAHint          string `json:"ja_hint"`
	CorrectSpelling string `json:"correct_spelling"`
	Explanation     string `json:"explanation"`
}

type SpellingEnvelope struct {
	Prompts []SpellingPrompt `json:"prompts"`
}

func validateSpelling(raw []byte) error {
	var env SpellingEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("invalid spelling JSON: %w", err)
	}
	if len(env.Prompts) != 5 {
		return fmt.Errorf("spelling prompts must be 5, got %d", len(env.Prompts))
	}
	return nil
}

type ListeningItem struct {
	Prompt      string   `json:"prompt"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Transcript  string   `json:"transcript"`
}

type ListeningEnvelope struct {
	Audio []ListeningItem `json:"audio"`
}

func validateListening(raw []byte) error {
	var env ListeningEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("invalid listening JSON: %w", err)
	}
	if len(env.Audio) != 5 {
		return fmt.Errorf("listening audio items must be 5, got %d", len(env.Audio))
	}
	for i, a := range env.Audio {
		if err := validateOptions(a.Options, a.AnswerIndex); err != nil {
			return fmt.Errorf("listening item %d: %w", i, err)
		}
	}
	return nil
}

func validateOptions(opts []string, answer int) error {
	if len(opts) != 4 {
		return fmt.Errorf("options must be 4, got %d", len(opts))
	}
	if answer < 0 || answer >= len(opts) {
		return fmt.Errorf("answer_index out of range: %d", answer)
	}
	return nil
}
