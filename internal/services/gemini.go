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
    {"npc_reply": "string"},
    {"npc_reply": "string"},
    {"npc_reply": "string"},
    {"npc_reply": "string"},
    {"npc_reply": "string"}
  ]
}
# IMPORTANT: The "turns" array MUST contain exactly 5 objects. Return ONLY the JSON object above (no explanation, no markdown, no surrounding text).`
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

	// The Gemini API might return markdown or surrounding text, so extract the first
	// well-formed JSON object using a simple brace-matching parser that handles
	// string literals and escapes.
	jsonString := string(contentBytes)
	extracted, ok := findJSONBlock(jsonString)
	if !ok {
		return QuestionPayload{}, fmt.Errorf("could not extract JSON from Gemini API response: %s", jsonString)
	}

	extractedJSON := []byte(extracted)

	// If mode is tavern and the extracted JSON doesn't have 5 turns, scan all JSON
	// blocks in the response and pick the first one that does.
	if mode == ModeTavern {
		var te TavernEnvelope
		if err := json.Unmarshal(extractedJSON, &te); err != nil || len(te.Turns) != 5 {
			blocks := findJSONBlocks(jsonString)
			for _, b := range blocks {
				var cand TavernEnvelope
				if err := json.Unmarshal([]byte(b), &cand); err == nil && len(cand.Turns) == 5 {
					extractedJSON = []byte(b)
					break
				}
			}
		}
	}

	return QuestionPayload{Mode: mode, Content: extractedJSON}, nil
}

// findJSONBlocks returns all top-level JSON objects found in s, in order.
func findJSONBlocks(s string) []string {
	results := []string{}
	inString := false
	escape := false
	depth := 0
	start := -1
	for i, ch := range s {
		if escape {
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
			continue
		}
		if ch == '}' {
			if depth > 0 {
				depth--
				if depth == 0 && start != -1 {
					results = append(results, s[start:i+1])
					start = -1
				}
			}
		}
	}
	return results
}

// findJSONBlock attempts to find the first top-level JSON object in s and returns it.
// It handles string literals and escaped quotes so braces inside strings are ignored.
func findJSONBlock(s string) (string, bool) {
	inString := false
	escape := false
	depth := 0
	start := -1
	for i, ch := range s {
		if escape {
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
			continue
		}
		if ch == '}' {
			if depth > 0 {
				depth--
				if depth == 0 && start != -1 {
					return s[start : i+1], true
				}
			}
		}
	}
	return "", false
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

// --- Batch evaluation types and functions ---

type TavernEvaluation struct {
	Outcome string `json:"outcome"` // "success"|"normal"|"fail"
	Reason  string `json:"reason"`
}

type batchEvalEnvelope struct {
	Evaluations []TavernEvaluation `json:"evaluations"`
}

// BatchEvaluateTavern evaluates 5 turns in one request.
// langPref: "en", "ja", or "both"
func (gc *GeminiClient) BatchEvaluateTavern(ctx context.Context, rubric []string, npcOpening string, npcReplies []TavernTurn, playerUtterances []string, langPref string) ([]TavernEvaluation, error) {
	if len(npcReplies) != 5 || len(playerUtterances) != 5 {
		return nil, fmt.Errorf("expected 5 npcReplies and 5 playerUtterances")
	}

	prompt := buildBatchEvalPrompt(rubric, npcOpening, npcReplies, playerUtterances, langPref)

	resp, err := gc.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fallbackEvaluations(fmt.Errorf("gemini generate error: %w", err)), nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return fallbackEvaluations(fmt.Errorf("no content from gemini")), nil
	}

	var contentBytes []byte
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			contentBytes = append(contentBytes, []byte(text)...)
		}
	}

	jsonStr := string(contentBytes)
	extracted, ok := findJSONBlock(jsonStr)
	if !ok {
		return fallbackEvaluations(fmt.Errorf("could not extract JSON from response: %s", jsonStr)), nil
	}

	var env batchEvalEnvelope
	if err := json.Unmarshal([]byte(extracted), &env); err != nil {
		return fallbackEvaluations(fmt.Errorf("invalid JSON: %w", err)), nil
	}

	if len(env.Evaluations) != 5 {
		return fallbackEvaluations(fmt.Errorf("evaluations length != 5: %d", len(env.Evaluations))), nil
	}

	for i := range env.Evaluations {
		switch env.Evaluations[i].Outcome {
		case "success", "normal", "fail":
			// ok
		default:
			return fallbackEvaluations(fmt.Errorf("invalid outcome: %s", env.Evaluations[i].Outcome)), nil
		}
	}

	return env.Evaluations, nil
}

func fallbackEvaluations(err error) []TavernEvaluation {
	fmt.Fprintf(os.Stderr, "BatchEvaluateTavern fallback: %v\n", err)
	res := make([]TavernEvaluation, 5)
	for i := range res {
		res[i] = TavernEvaluation{
			Outcome: "normal",
			Reason:  "Evaluation failed — defaulted to normal",
		}
	}
	return res
}

func buildBatchEvalPrompt(rubric []string, npcOpening string, npcReplies []TavernTurn, playerUtterances []string, langPref string) string {
	var b strings.Builder
	b.WriteString("You are an English conversation evaluator. Use the evaluation rubric below to judge each player utterance for the corresponding NPC reply. ")
	b.WriteString("Return only valid JSON with this format:\n")
	b.WriteString(`{"evaluations":[{"outcome":"success|normal|fail","reason":"short reason in the chosen language"}, ...]}` + "\n\n")

	b.WriteString("Evaluation rubric (EN):\n")
	for i, r := range rubric {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, r))
	}
	b.WriteString("\n評価基準（日本語）:\n")
	for i, r := range rubric {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, r))
	}
	b.WriteString("\n")

	b.WriteString("NPC Opening:\n")
	b.WriteString(npcOpening + "\n\n")

	for i := 0; i < 5; i++ {
		b.WriteString(fmt.Sprintf("Turn %d NPC reply (EN):\n%s\n", i+1, npcReplies[i].NPCReply))
		b.WriteString(fmt.Sprintf("Turn %d Player utterance:\n%s\n\n", i+1, playerUtterances[i]))
	}

	if langPref == "ja" {
		b.WriteString("Please return reasons in Japanese.\n")
	} else if langPref == "en" {
		b.WriteString("Please return reasons in English.\n")
	} else {
		b.WriteString("Please return reasons in both English and Japanese (e.g. 'EN reason.／JP reason.').\n")
	}

	b.WriteString("Make sure the JSON is valid and contains exactly 5 evaluations in order.\n")
	return b.String()
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
