# AI Challenge Security Fix - Template-Based System

## Summary

Fixed the security vulnerability where hardcoded answers allowed bypassing AI challenges.
The new system uses parameterized templates and quality-based evaluation scoring.

## What Was Changed

### Old System (Vulnerable)
- ~110 hardcoded Q&A pairs with hashed answers
- Answers could be reversed due to small question pool
- Single correct answer per question
- Easy to exploit by reverse-engineering answers

### New System (Secure)
- **Template-based generation**: 17 questions with VRF-selected parameters
- **Quality-based scoring**: Evaluates response quality instead of exact match
- **No hardcoded answers**: Questions generated dynamically
- **Anti-cheat detection**: Flags identical copy-paste responses

---

## Implementation Details

### Templates
Questions use placeholders like `{algorithm}`, `{blockchain}`, `{operation}` that are
filled with VRF-selected values at challenge generation time.

### Scoring System
Quality-based evaluation (max 100 points):
1. **Length Score (35 pts max)**: Longer responses indicate real AI usage
2. **Keyword Score (35 pts max)**: Technical terms relevant to category
3. **Format Score (30 pts max)**: Structure, numbers, bullet points

### Categories
- algorithms, data_structures, blockchain, math, networking, databases
- security, cryptography, machine_learning

### Anti-Cheat
- Responses with 3+ identical copies flagged as cheaters
- Slashing penalty: 20% of stake + reputation loss

---

## Files Changed

- `x/agent/keeper/challenge.go` - Main implementation with templates and scoring
- `x/agent/keeper/challenge_test.go` - Quality scoring tests
- `x/agent/keeper/export_test.go` - Test helpers for new functions
- `AI_CHALLENGE_SECURITY.md` - This documentation

## Security Improvements

1. **No reverse-engineering**: No answers in source code
2. **Dynamic questions**: VRF-based generation
3. **Quality validation**: Rewards genuine AI responses
4. **Anti-cheat**: Detects copy-paste patterns

---

## Security Issue Details

### Problem
Running `grep -r "Answer" x/agent/keeper/` reveals all AI Challenge answers are stored in plaintext in the open-source codebase.

This allows anyone to pass challenges without running any AI model.

### Root Cause
The `Answer` field was stored directly in `QuestionBank`:
```go
var QuestionBank = []Challenge{
    {Question: "Capital of France?", Answer: "paris"},
}
```

### Solution
Three layers implemented:

1. **Remove hardcoded answers** — delete Answer field from questions
2. **VRF/seed generation** — generate questions from block hash, unpredictable
3. **Quality scoring** — validators judged by response quality, not exact match