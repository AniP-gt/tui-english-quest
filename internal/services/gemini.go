package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

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

// FetchQuestions fetches five questions for the given mode.
func FetchQuestions(ctx context.Context, mode string) (QuestionPayload, error) {
	_ = ctx
	// TODO: Call Gemini API. For now, return sample payload.
	// In a real implementation, this would involve HTTP requests and API key handling.
	// Example:
	// resp, err := http.Get(fmt.Sprintf("https://api.gemini.com/v1/questions?mode=%s", mode))
	// if err != nil {
	//     return QuestionPayload{}, fmt.Errorf("failed to fetch from Gemini API: %w", err)
	// }
	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	//     return QuestionPayload{}, fmt.Errorf("failed to read Gemini API response: %w", err)
	// }
	// return QuestionPayload{Mode: mode, Content: body}, nil
	return QuestionPayload{Mode: mode, Content: samplePayload(mode)}, nil
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

type vocabQuestion struct {
	EnemyName   string   `json:"enemy_name"`
	Word        string   `json:"word"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Explanation string   `json:"explanation"`
}

type vocabEnvelope struct {
	Questions []vocabQuestion `json:"questions"`
}

func validateVocab(raw []byte) error {
	var env vocabEnvelope
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

type grammarTrap struct {
	TrapName    string   `json:"trap_name"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Explanation string   `json:"explanation"`
}

type grammarEnvelope struct {
	Traps []grammarTrap `json:"traps"`
}

func validateGrammar(raw []byte) error {
	var env grammarEnvelope
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

type tavernTurn struct {
	NPCReply string `json:"npc_reply"`
}

type tavernEnvelope struct {
	NPCName          string       `json:"npc_name"`
	NPCOpening       string       `json:"npc_opening"`
	EvaluationRubric []string     `json:"evaluation_rubric"`
	Turns            []tavernTurn `json:"turns"`
}

func validateTavern(raw []byte) error {
	var env tavernEnvelope
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

type spellingPrompt struct {
	JAHint          string `json:"ja_hint"`
	CorrectSpelling string `json:"correct_spelling"`
	Explanation     string `json:"explanation"`
}

type spellingEnvelope struct {
	Prompts []spellingPrompt `json:"prompts"`
}

func validateSpelling(raw []byte) error {
	var env spellingEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("invalid spelling JSON: %w", err)
	}
	if len(env.Prompts) != 5 {
		return fmt.Errorf("spelling prompts must be 5, got %d", len(env.Prompts))
	}
	return nil
}

type listeningItem struct {
	Prompt      string   `json:"prompt"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Transcript  string   `json:"transcript"`
}

type listeningEnvelope struct {
	Audio []listeningItem `json:"audio"`
}

func validateListening(raw []byte) error {
	var env listeningEnvelope
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

// samplePayload returns a minimal valid JSON payload per mode for offline runs.
func samplePayload(mode string) []byte {
	switch mode {
	case ModeVocab:
		return []byte(`{"questions":[{"enemy_name":"Slime","word":"maintain","options":["to keep","to break","to borrow","to throw"],"answer_index":0,"explanation":"to maintain means to keep something in good condition"},{"enemy_name":"Slime","word":"reduce","options":["to lower","to increase","to borrow","to throw"],"answer_index":0,"explanation":"reduce means make smaller"},{"enemy_name":"Slime","word":"create","options":["to make","to delete","to borrow","to throw"],"answer_index":0,"explanation":"create means make"},{"enemy_name":"Slime","word":"borrow","options":["to take and return","to throw","to buy","to sell"],"answer_index":0,"explanation":"borrow means take temporarily"},{"enemy_name":"Slime","word":"throw","options":["to toss","to keep","to borrow","to heal"],"answer_index":0,"explanation":"throw means toss"}]}`)
	case ModeGrammar:
		return []byte(`{"traps":[{"trap_name":"Past Tense Trap","question":"Which sentence is correct?","options":["I go yesterday.","I went yesterday.","I gone yesterday.","I going yesterday."],"answer_index":1,"explanation":"Past tense of go is went."},{"trap_name":"Do/Does","question":"Choose the correct sentence","options":["He don't like tea.","He doesn't like tea.","He not like tea.","He likes not tea."],"answer_index":1,"explanation":"Use doesn't + base verb."},{"trap_name":"Preposition","question":"Choose the correct preposition","options":["on Monday","in Monday","at Monday","by Monday"],"answer_index":0,"explanation":"Use on with days."},{"trap_name":"Article","question":"Choose the correct article","options":["I saw a moon","I saw an moon","I saw the moon","I saw moon"],"answer_index":2,"explanation":"Use the for the moon."},{"trap_name":"Verb Form","question":"Choose the correct form","options":["She go to work","She goes to work","She going to work","She gone to work"],"answer_index":1,"explanation":"Third person singular adds -es."}]}`)
	case ModeTavern:
		return []byte(`{"npc_name":"Old Jaro","npc_opening":"Hey traveler, what brings you here tonight?","evaluation_rubric":["Success: fluent, relevant, task completed","Normal: understandable, minor issues","Fail: unclear or off-topic"],"turns":[{"npc_reply":"The shop is down that road."},{"npc_reply":"The inn is upstairs."},{"npc_reply":"The tavern serves stew."},{"npc_reply":"The guard is outside."},{"npc_reply":"Travel safe, friend."}]}`)
	case ModeSpelling:
		return []byte(`{"prompts":[{"ja_hint":"維持する","correct_spelling":"maintain","explanation":"main + tain"},{"ja_hint":"減らす","correct_spelling":"reduce","explanation":"re + duce"},{"ja_hint":"作る","correct_spelling":"create","explanation":"create"},{"ja_hint":"借りる","correct_spelling":"borrow","explanation":"borrow"},{"ja_hint":"投げる","correct_spelling":"throw","explanation":"throw"}]}`)
	case ModeListening:
		return []byte(`{"audio":[{"prompt":"What does she want to buy?","options":["Shoes","Coffee","A book","Food"],"answer_index":1,"transcript":"I'm going to buy some coffee."},{"prompt":"Where is the shop?","options":["Down the road","In the sky","Underwater","On the roof"],"answer_index":0,"transcript":"The shop is down the road."},{"prompt":"What time is it?","options":["Morning","Noon","Evening","Night"],"answer_index":2,"transcript":"It's evening now."},{"prompt":"How many items?","options":["One","Two","Three","Four"],"answer_index":1,"transcript":"I will take two."},{"prompt":"What does he drink?","options":["Water","Juice","Tea","Soda"],"answer_index":2,"transcript":"He likes tea."}]}`)
	default:
		return []byte("[]")
	}
}
