# Gemini Contracts - TUI English Quest v2.0

All responses MUST be valid JSON. Fetch exactly 5 items per session; reject/redo if count <5 or JSON invalid.

## Vocabulary Battle
```json
{
  "questions": [
    {
      "enemy_name": "Slime",
      "word": "maintain",
      "options": ["to keep", "to break", "to borrow", "to throw"],
      "answer_index": 0,
      "explanation": "to maintain means to keep something in good condition"
    }
  ]
}
```

## Grammar Dungeon
```json
{
  "traps": [
    {
      "trap_name": "Past Tense Trap",
      "question": "Which sentence is correct?",
      "options": ["I go yesterday.", "I went yesterday.", "I gone yesterday.", "I going yesterday."],
      "answer_index": 1,
      "explanation": "Past tense of go is went."
    }
  ]
}
```

## Conversation Tavern
```json
{
  "npc_name": "Old Jaro",
  "npc_opening": "Hey traveler, what brings you here tonight?",
  "evaluation_rubric": [
    "Success: fluent, relevant, task completed",
    "Normal: understandable, minor issues",
    "Fail: unclear or off-topic"
  ],
  "turns": [
    {"npc_reply": "If you need supplies, the shop is down that road."}
  ]
}
```

## Spelling Challenge
```json
{
  "prompts": [
    {
      "ja_hint": "維持する",
      "correct_spelling": "maintain",
      "explanation": "main + tain pattern"
    }
  ]
}
```

## Listening Cave
```json
{
  "audio": [
    {
      "prompt": "What does she want to buy?",
      "options": ["Shoes", "Coffee", "A book", "Food"],
      "answer_index": 1,
      "transcript": "I’m going to buy some coffee."
    }
  ]
}
```

## AI Weakness Analysis
```json
{
  "weak_points": ["past tense", "spelling"],
  "strength_points": ["basic vocabulary"],
  "recommendation": "Play Spelling Challenge twice today."
}
```
