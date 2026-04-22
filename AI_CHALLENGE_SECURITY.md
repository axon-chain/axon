# AI Challenge Security Fix - Template-Based System

## Summary

Fixed the security vulnerability where hardcoded answers allowed bypassing AI challenges.
The new system uses parameterized templates and cross-evaluation scoring.

## What Was Changed

### Old System (Vulnerable)
- ~110 hardcoded Q&A pairs with hashed answers
- Answers could be reversed due to small question pool
- Single correct answer per question

### New System (Secure)
- **Template-based generation**: Questions generated from parameterized templates
- **VRF-based randomness**: Random values selected from pools for each placeholder
- **Cross-evaluation**: Validators score each other's answers based on consensus
- No hardcoded answers in source

---

## Implementation Details

### Templates
Questions use placeholders like `{algorithm}`, `{blockchain}`, `{operation}` that are
filled with VRF-selected values at challenge generation time.

### Categories
- algorithms, blockchain, math, networking, databases, security, cryptography, machine_learning

### Scoring
- Unique answer: 90 points
- 2 validators same answer: 70 points  
- 3+ identical (cheating): penalized
- Cross-validation replaces hash comparison

---

## Files Changed

- `x/agent/keeper/challenge.go` - Main implementation
- `x/agent/keeper/challenge_test.go` - Updated tests
- `x/agent/keeper/export_test.go` - Test helpers
4. Est-ce que les reponses sont comparees on-chain ou off-chain?

Fichiers a regarder: x/agent/keeper/questions.go, x/agent/keeper/challenge.go
```

### Question 2: Proposer une solution

```
Je veux proposer une amelioration pour securiser le AI Challenge system.

Probleme: Les reponses sont hardcodees dans le code open-source.

Solution proposee:
1. Supprimer le champ "Answer" des questions
2. Generer les questions via VRF/seed depuis le block hash
3. Pour les questions deterministes (math, logique), calculer la reponse on-chain

Peux-tu ecrire le code Go pour:
- SupprimerAnswer des structures
- Ajouter une fonction GenerateChallengeFromSeed()
- Modifier la verification pour calculer() au lieu de comparer?

Fichier: x/agent/keeper/
```

### Question 3: Cross-scoring

```
Pour les questions ouvertes (resume, code review), comment implementer un systeme de scoring distribue?

Ideas:
- Les validateurs notent les reponses des autres
- Penalite si >40% des reponses sont quasi-identiques (detection de copie)
- Score final = mediane des notes recues

Questions:
1. Comment collecter les scores entre validateurs?
2. Quelle fonction detecte les reponses copiees?
3. Comment appliquer les penalites?

Code actuel a modifier: x/agent/keeper/challenge_scoring.go
```

---

## Template pour Issue GitHub

Titre:
```
[Security] AI Challenge answers exposed in open-source — breaks integrity
```

Corps:
```markdown
## Problem

Running `grep -r "Answer" x/agent/keeper/` reveals all AI Challenge answers are stored in plaintext in the open-source codebase.

This allows anyone to pass challenges without running any AI model.

## Root Cause

The `Answer` field is stored directly in `QuestionBank`:
```go
var QuestionBank = []Challenge{
    {Question: "Capital of France?", Answer: "paris"},
}
```

## Solution

Three layers needed:

1. **Remove hardcoded answers** — delete Answer field from questions
2. **VRF/seed generation** — generate questions from block hash, unpredictable
3. **Cross-scoring** — validators rate each other's answers, detect copying

## Files to Modify

- `x/agent/keeper/questions.go` — remove Answer field
- `x/agent/keeper/challenge_vrf.go` — new VRF-based generation  
- `x/agent/keeper/challenge_scoring.go` — distributed scoring
```

---

## Prochaines Etapes

1. **Confirmer le bug** — Poser question 1 a Claude
2. **Ecrire le fix** — Poser question 2 a Claude  
3. **Creer l'issue** — Poster sur https://github.com/axon-chain/axon/issues
4. **Soumettre PR** — Proposer les corrections

---

C'est un excellent premier probleme a contribuer car:
- Impact securite evident
- Solution technique interessante (VRF, cross-scoring)
- Facile a expliquer et review