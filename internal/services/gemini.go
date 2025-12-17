package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings" // Added strings import
	"tui-english-quest/internal/config"

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

// FetchQuestions fetches questions for the given mode using the Gemini API.
func FetchQuestions(ctx context.Context, mode string) (QuestionPayload, error) {
	gc, err := NewGeminiClient(ctx)
	if err != nil {
		return QuestionPayload{}, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	// The genai.GenerativeModel does not have a Close() method, so we don't defer it here.

	// Read user language preference and configured questions per session
	langPref := "en"
	cfg, _ := config.LoadConfig()
	if cfg.LangPref == "ja" {
		langPref = "ja"
	}
	N := 5
	if cfg.QuestionsPerSession > 0 {
		N = cfg.QuestionsPerSession
	}

	prompt := fmt.Sprintf("Generate %d *new and diverse* %s questions in JSON format. The JSON should strictly adhere to the following structure for %s mode:\n\n", N, mode, mode)

	// Language instruction: enforce English for problem texts in specific modes
	switch mode {
	case ModeGrammar, ModeTavern, ModeListening:
		// Problem text (questions, NPC replies, listening prompts and options) must be English.
		if langPref == "ja" {
			prompt = "Write all problem texts, prompts, NPC replies and options in English. Provide explanations/transcripts/evaluation reasons in Japanese. Return only JSON.\n\n" + prompt
		} else {
			prompt = "Write problem texts, prompts, NPC replies, options and explanations in English. Return only JSON.\n\n" + prompt
		}
	default:
		// Vocab and Spelling follow user langPref for problem text.
		if langPref == "ja" {
			prompt = "Please write the problem text in Japanese, but provide answers/options in English. Return only JSON.\n\n" + prompt
		} else {
			prompt = "Please write the problem text and answers in English. Return only JSON.\n\n" + prompt
		}
	}

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
Return ONLY the JSON object above (no explanation, no markdown, no surrounding text).`
		// Add explicit constraint about turns count matching N
		prompt += fmt.Sprintf("\n# IMPORTANT: The \"turns\" array MUST contain exactly %d objects. Return ONLY the JSON object above (no explanation, no markdown, no surrounding text).\n", N)
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

	// If mode is tavern and the extracted JSON doesn't have the expected number
	// of turns, scan all JSON blocks in the response and pick the first one that does.
	if mode == ModeTavern {
		var te TavernEnvelope
		if err := json.Unmarshal(extractedJSON, &te); err != nil || len(te.Turns) != N {
			blocks := findJSONBlocks(jsonString)
			for _, b := range blocks {
				var cand TavernEnvelope
				if err := json.Unmarshal([]byte(b), &cand); err == nil && len(cand.Turns) == N {
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

// ValidatePayload validates JSON shape and count (uses configured QuestionsPerSession).
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
	cfg, _ := config.LoadConfig()
	N := 5
	if cfg.QuestionsPerSession > 0 {
		N = cfg.QuestionsPerSession
	}
	if len(env.Questions) != N {
		return fmt.Errorf("vocab questions must be %d, got %d", N, len(env.Questions))
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
	cfg, _ := config.LoadConfig()
	N := 5
	if cfg.QuestionsPerSession > 0 {
		N = cfg.QuestionsPerSession
	}
	if len(env.Traps) != N {
		return fmt.Errorf("grammar traps must be %d, got %d", N, len(env.Traps))
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
	cfg, _ := config.LoadConfig()
	N := 5
	if cfg.QuestionsPerSession > 0 {
		N = cfg.QuestionsPerSession
	}
	if len(env.Turns) != N {
		return fmt.Errorf("tavern turns must be %d, got %d", N, len(env.Turns))
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

// BatchEvaluateTavern evaluates N turns in one request.
// langPref: "en" or "ja"
func (gc *GeminiClient) BatchEvaluateTavern(ctx context.Context, rubric []string, npcOpening string, npcReplies []TavernTurn, playerUtterances []string, langPref string) ([]TavernEvaluation, error) {
	cfg, _ := config.LoadConfig()
	N := 5
	if cfg.QuestionsPerSession > 0 {
		N = cfg.QuestionsPerSession
	}
	if len(npcReplies) != N || len(playerUtterances) != N {
		return nil, fmt.Errorf("expected %d npcReplies and %d playerUtterances", N, N)
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

	expected := len(npcReplies)
	if len(env.Evaluations) != expected {
		return fallbackEvaluations(fmt.Errorf("evaluations length != %d: %d", expected, len(env.Evaluations))), nil
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

	count := len(npcReplies)
	for i := 0; i < count; i++ {
		b.WriteString(fmt.Sprintf("Turn %d NPC reply (EN):\n%s\n", i+1, npcReplies[i].NPCReply))
		b.WriteString(fmt.Sprintf("Turn %d Player utterance:\n%s\n\n", i+1, playerUtterances[i]))
	}

	if langPref == "ja" {
		b.WriteString("Please return reasons in Japanese.\n")
	} else if langPref == "en" {
		b.WriteString("Please return reasons in English.\n")
	} else {
		b.WriteString("Please return reasons in English (brief).\n")
	}

	b.WriteString(fmt.Sprintf("Make sure the JSON is valid and contains exactly %d evaluations in order.\n", count))
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
	cfg, _ := config.LoadConfig()
	N := 5
	if cfg.QuestionsPerSession > 0 {
		N = cfg.QuestionsPerSession
	}
	if len(env.Prompts) != N {
		return fmt.Errorf("spelling prompts must be %d, got %d", N, len(env.Prompts))
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
	cfg, _ := config.LoadConfig()
	N := 5
	if cfg.QuestionsPerSession > 0 {
		N = cfg.QuestionsPerSession
	}
	if len(env.Audio) != N {
		return fmt.Errorf("listening audio items must be %d, got %d", N, len(env.Audio))
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
