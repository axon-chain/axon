package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

// L2 report storage prefixes
const (
	L2ReportKeyPrefix      = "L2Report/"
	L2ReportIndexKeyPrefix = "L2ReportIdx/"
	L2ReportCountPrefix    = "L2ReportCnt/"

	DefaultL2BudgetPerAgentMillis int64 = 100     // 0.1 per agent
	DefaultL2BudgetCapMillis      int64 = 100_000 // 100.0
	L2MaxReportsPerEpoch          int64 = 50

	EvidenceWeightValid int64 = 1000 // ×1.0
)

type L2Report struct {
	Reporter string
	Target   string
	Score    int8   // +1 or -1
	Evidence string // tx hash hex, empty = no evidence
	Reason   string // human-readable justification
	Weight   int64  // computed weight in millis (×1000)
	Epoch    uint64
}

func l2ReportKey(epoch uint64, reporter, target string) []byte {
	key := []byte(L2ReportKeyPrefix)
	key = append(key, types.Uint64ToBytes(epoch)...)
	key = append(key, []byte("/"+reporter+"/"+target)...)
	return key
}

func l2ReportTargetPrefix(epoch uint64, target string) []byte {
	key := []byte(L2ReportIndexKeyPrefix)
	key = append(key, types.Uint64ToBytes(epoch)...)
	key = append(key, []byte("/"+target+"/")...)
	return key
}

func l2ReportTargetKey(epoch uint64, target, reporter string) []byte {
	key := l2ReportTargetPrefix(epoch, target)
	key = append(key, []byte(reporter)...)
	return key
}

func l2ReporterCountKey(epoch uint64, reporter string) []byte {
	key := []byte(L2ReportCountPrefix)
	key = append(key, types.Uint64ToBytes(epoch)...)
	key = append(key, []byte("/"+reporter)...)
	return key
}

func (k Keeper) HasL2Report(ctx sdk.Context, epoch uint64, reporter, target string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(l2ReportKey(epoch, reporter, target))
}

func (k Keeper) GetL2ReporterEpochCount(ctx sdk.Context, epoch uint64, reporter string) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(l2ReporterCountKey(epoch, reporter))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) incrL2ReporterCount(ctx sdk.Context, epoch uint64, reporter string) {
	cnt := k.GetL2ReporterEpochCount(ctx, epoch, reporter) + 1
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(cnt))
	store.Set(l2ReporterCountKey(epoch, reporter), bz)
}

// GetL2ReportStats returns (positiveCount, negativeCount, uniqueReporters) for a target in a given epoch.
func (k Keeper) GetL2ReportStats(ctx sdk.Context, epoch uint64, target string) (uint64, uint64, uint64) {
	store := ctx.KVStore(k.storeKey)
	prefix := l2ReportTargetPrefix(epoch, target)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var posCount, negCount, unique uint64
	for ; iterator.Valid(); iterator.Next() {
		score, _ := decodeL2Report(iterator.Value())
		if score > 0 {
			posCount++
		} else {
			negCount++
		}
		unique++
	}
	return posCount, negCount, unique
}

// L2ReportDetail contains the decoded fields of a stored L2 report.
type L2ReportDetail struct {
	Reporter string
	Score    int8
	Weight   int64
	Reason   string
}

// GetL2ReportDetails returns all report details for a target in a given epoch,
// including the reason field. Useful for query/audit endpoints.
func (k Keeper) GetL2ReportDetails(ctx sdk.Context, epoch uint64, target string) []L2ReportDetail {
	store := ctx.KVStore(k.storeKey)
	prefix := l2ReportTargetPrefix(epoch, target)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var details []L2ReportDetail
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		reporter := key[len(string(prefix)):]
		score, weight, reason := decodeL2ReportFull(iterator.Value())
		details = append(details, L2ReportDetail{
			Reporter: reporter,
			Score:    score,
			Weight:   weight,
			Reason:   reason,
		})
	}
	return details
}

// GetL2TargetReportCount returns the number of reports targeting this address in the given epoch.
func (k Keeper) GetL2TargetReportCount(ctx sdk.Context, epoch uint64, target string) int64 {
	store := ctx.KVStore(k.storeKey)
	prefix := l2ReportTargetPrefix(epoch, target)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	count := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

// SubmitL2Report validates and stores a reputation report.
func (k Keeper) SubmitL2Report(ctx sdk.Context, reporter, target string, score int8, evidence, reason string) error {
	params := k.GetParams(ctx)

	if score != 1 && score != -1 {
		return fmt.Errorf("invalid score: must be +1 or -1")
	}
	if reporter == target {
		return fmt.Errorf("cannot self-report")
	}

	reporterAgent, found := k.GetAgent(ctx, reporter)
	if !found {
		return fmt.Errorf("caller not registered")
	}
	reporterRep := k.GetTotalReputation(ctx, reporter)
	if reporterRep < l2MinReporterRepMillis(params) {
		return fmt.Errorf("reputation too low")
	}
	if ctx.BlockHeight()-reporterAgent.RegisteredAt < params.L2MinAccountAge {
		return fmt.Errorf("account too new")
	}

	targetAgent, found := k.GetAgent(ctx, target)
	if !found {
		return fmt.Errorf("target not registered")
	}
	if k.IsV111UpgradeActivated(ctx) && targetAgent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
		return fmt.Errorf("target is deregistering")
	}

	epoch := k.GetCurrentEpoch(ctx)
	if k.HasL2Report(ctx, epoch, reporter, target) {
		return fmt.Errorf("already reported")
	}
	if k.GetL2ReporterEpochCount(ctx, epoch, reporter) >= L2MaxReportsPerEpoch {
		return fmt.Errorf("epoch report limit reached (%d)", L2MaxReportsPerEpoch)
	}

	// Compute base weight from reporter's reputation using deterministic LegacyDec sqrt
	repFrac := math.LegacyNewDec(reporterRep).Quo(math.LegacyNewDec(100_000))
	baseWeightDec := ApproxSqrt(repFrac).MulInt64(1000)
	baseWeight := baseWeightDec.TruncateInt64()

	evidenceWeight := l2NoEvidenceWeightMillis(params)
	if !k.IsV110UpgradeActivated(ctx) {
		// Pre-upgrade: preserve exact v1.0.0 behavior — any non-empty string grants full weight.
		// Do NOT normalize here; normalization could turn whitespace-only evidence into ""
		// and produce a different weight, breaking historical replay consensus.
		if evidence != "" {
			evidenceWeight = EvidenceWeightValid
		}
	} else {
		evidence = normalizeEvidenceForStorage(evidence)
		if isValidEvidenceFormat(evidence) && k.HasEvidenceTxHash(ctx, evidence) {
			evidenceWeight = EvidenceWeightValid
		}
	}

	weight := baseWeight * evidenceWeight / 1000

	if len(reason) > 256 {
		reason = reason[:256]
	}

	report := L2Report{
		Reporter: reporter,
		Target:   target,
		Score:    score,
		Evidence: evidence,
		Reason:   reason,
		Weight:   weight,
		Epoch:    epoch,
	}

	k.storeL2Report(ctx, report)
	k.incrL2ReporterCount(ctx, epoch, reporter)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"l2_report_submitted",
		sdk.NewAttribute("reporter", reporter),
		sdk.NewAttribute("target", target),
		sdk.NewAttribute("score", fmt.Sprintf("%d", score)),
		sdk.NewAttribute("weight", fmt.Sprintf("%d", weight)),
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("reason", reason),
	))

	return nil
}

func isValidEvidenceFormat(evidence string) bool {
	normalized, ok := normalizeEvidenceHash(evidence)
	if !ok {
		return false
	}
	decoded, err := hex.DecodeString(normalized)
	return err == nil && len(decoded) == commonHashLength
}

const commonHashLength = 32

func normalizeEvidenceForStorage(evidence string) string {
	normalized, ok := normalizeEvidenceHash(evidence)
	if !ok {
		return strings.TrimSpace(evidence)
	}
	return normalized
}

func normalizeEvidenceHash(evidence string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(evidence))
	normalized = strings.TrimPrefix(normalized, "0x")
	if len(normalized) != commonHashLength*2 {
		return "", false
	}
	return normalized, true
}

func (k Keeper) storeL2Report(ctx sdk.Context, r L2Report) {
	store := ctx.KVStore(k.storeKey)

	// Primary key: epoch/reporter/target
	value := encodeL2Report(r)
	store.Set(l2ReportKey(r.Epoch, r.Reporter, r.Target), value)

	// Reverse index: epoch/target/reporter (for iterating reports about a target)
	store.Set(l2ReportTargetKey(r.Epoch, r.Target, r.Reporter), value)
}

// Wire format: [score:1][weight:8][reasonLen:2][reason:N]
func encodeL2Report(r L2Report) []byte {
	scoreByte := byte(0)
	if r.Score > 0 {
		scoreByte = 1
	}
	reasonBytes := []byte(r.Reason)
	if len(reasonBytes) > 256 {
		reasonBytes = reasonBytes[:256]
	}
	bz := make([]byte, 11+len(reasonBytes))
	bz[0] = scoreByte
	binary.BigEndian.PutUint64(bz[1:9], uint64(r.Weight))
	binary.BigEndian.PutUint16(bz[9:11], uint16(len(reasonBytes)))
	copy(bz[11:], reasonBytes)
	return bz
}

func decodeL2Report(bz []byte) (score int8, weight int64) {
	if len(bz) < 9 {
		return 0, 0
	}
	if bz[0] == 1 {
		score = 1
	} else {
		score = -1
	}
	weight = int64(binary.BigEndian.Uint64(bz[1:]))
	return
}

func decodeL2ReportFull(bz []byte) (score int8, weight int64, reason string) {
	if len(bz) < 9 {
		return 0, 0, ""
	}
	if bz[0] == 1 {
		score = 1
	} else {
		score = -1
	}
	weight = int64(binary.BigEndian.Uint64(bz[1:9]))
	if len(bz) >= 11 {
		reasonLen := int(binary.BigEndian.Uint16(bz[9:11]))
		if len(bz) >= 11+reasonLen {
			reason = string(bz[11 : 11+reasonLen])
		}
	}
	return
}

// SettleL2Reputation processes all L2 reports for the epoch.
// Implements M6 (anti-cheat), M7 (budget), M8 (full settlement flow).
// Reads L2 params from governance.
func (k Keeper) SettleL2Reputation(ctx sdk.Context, epoch uint64) {
	params := k.GetParams(ctx)

	rawDeltas := k.computeRawL2Deltas(ctx, epoch)

	agentCount := int64(0)
	k.IterateAgents(ctx, func(_ types.Agent) bool {
		agentCount++
		return false
	})

	budgetPerAgent := l2BudgetPerAgentMillis(params)
	budgetCap := l2BudgetCapMillis(params)

	budget := budgetPerAgent * agentCount
	if budget > budgetCap {
		budget = budgetCap
	}

	positiveSum := int64(0)
	for _, d := range rawDeltas {
		if d > 0 {
			positiveSum += d
		}
	}

	// Step 6: Apply scaled deltas using LegacyDec for deterministic precision
	budgetDec := math.LegacyNewDec(budget)
	positiveSumDec := math.LegacyNewDec(positiveSum)

	for addr, delta := range rawDeltas {
		scaled := delta
		if delta > 0 && positiveSum > budget {
			scaledDec := math.LegacyNewDec(delta).Mul(budgetDec).Quo(positiveSumDec)
			scaled = scaledDec.TruncateInt64()
		}

		current := k.GetL2Score(ctx, addr)
		newScore := current + scaled
		if newScore < 0 {
			newScore = 0
		}
		if newScore > k.l2CapMillis(ctx) {
			newScore = k.l2CapMillis(ctx)
		}
		k.SetL2Score(ctx, addr, newScore)
	}

	// Step 7: Cleanup old epoch data to prevent KV store growth
	if epoch > 2 {
		k.cleanupL2Reports(ctx, epoch-2)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"l2_reputation_settled",
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("agents_affected", fmt.Sprintf("%d", len(rawDeltas))),
		sdk.NewAttribute("budget_millis", fmt.Sprintf("%d", budget)),
		sdk.NewAttribute("positive_sum_millis", fmt.Sprintf("%d", positiveSum)),
	))
}

// computeRawL2Deltas collects all epoch reports, applies M6 anti-cheat, and computes per-target raw deltas.
func (k Keeper) computeRawL2Deltas(ctx sdk.Context, epoch uint64) map[string]int64 {
	params := k.GetParams(ctx)
	store := ctx.KVStore(k.storeKey)

	// Collect all reports this epoch
	type reportEntry struct {
		reporter string
		target   string
		score    int8
		weight   int64
	}

	var allReports []reportEntry
	prefix := []byte(L2ReportKeyPrefix)
	prefix = append(prefix, types.Uint64ToBytes(epoch)...)
	prefix = append(prefix, '/')

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		// Parse reporter and target from key
		// Key format after prefix: reporter/target
		rest := key[len(string(prefix)):]
		reporter, target := parseReporterTarget(rest)
		if reporter == "" || target == "" {
			continue
		}
		score, weight := decodeL2Report(iterator.Value())
		allReports = append(allReports, reportEntry{
			reporter: reporter, target: target, score: score, weight: weight,
		})
	}

	// M6 Rule 2: Spam detection — reporters with > 50 positive reports this epoch
	reporterPosCounts := make(map[string]int64)
	for _, r := range allReports {
		if r.score > 0 {
			reporterPosCounts[r.reporter]++
		}
	}
	spammers := make(map[string]bool)
	spamThreshold := l2AbuseThreshold(params)
	for addr, cnt := range reporterPosCounts {
		if cnt > spamThreshold {
			spammers[addr] = true
		}
	}

	// M6 Rule 1: Mutual rating detection
	reportSet := make(map[string]int8) // "A→B" → score
	for _, r := range allReports {
		reportSet[r.reporter+"→"+r.target] = r.score
	}

	// Apply adjustments and aggregate per-target
	targetWeightedSum := make(map[string]int64)
	targetWeightTotal := make(map[string]int64)
	mutualPenalty := l2MutualReportPenaltyMillis(params)

	for i := range allReports {
		r := &allReports[i]
		w := r.weight

		// M6 Rule 2: Spammer weight = 0
		if spammers[r.reporter] {
			w = 0
		}

		// M6 Rule 1: Mutual positive ratings → weight × 0.1
		if r.score > 0 {
			reverseKey := r.target + "→" + r.reporter
			if reverseScore, exists := reportSet[reverseKey]; exists && reverseScore > 0 {
				w = w * mutualPenalty / 1000
			}
		}

		if w <= 0 {
			continue
		}

		targetWeightedSum[r.target] += int64(r.score) * w
		targetWeightTotal[r.target] += w
	}

	// Compute raw_delta per target: weighted_sum / weight_total (range [-1000, +1000] millis)
	rawDeltas := make(map[string]int64, len(targetWeightedSum))
	for addr, wsum := range targetWeightedSum {
		wtotal := targetWeightTotal[addr]
		if wtotal <= 0 {
			continue
		}
		rawDeltas[addr] = wsum * 1000 / wtotal
	}

	return rawDeltas
}

func parseReporterTarget(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return s[:i], s[i+1:]
		}
	}
	return "", ""
}

// cleanupL2Reports removes all report data for the specified epoch from the KV store.
func (k Keeper) cleanupL2Reports(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.storeKey)

	prefixes := [][]byte{
		append([]byte(L2ReportKeyPrefix), types.Uint64ToBytes(epoch)...),
		append([]byte(L2ReportIndexKeyPrefix), types.Uint64ToBytes(epoch)...),
		append([]byte(L2ReportCountPrefix), types.Uint64ToBytes(epoch)...),
	}

	for _, prefix := range prefixes {
		iter := storetypes.KVStorePrefixIterator(store, prefix)
		keysToDelete := [][]byte{}
		for ; iter.Valid(); iter.Next() {
			keysToDelete = append(keysToDelete, iter.Key())
		}
		iter.Close()
		for _, key := range keysToDelete {
			store.Delete(key)
		}
	}
}

func l2MinReporterRepMillis(params types.Params) int64 {
	if params.L2MinReporterRep <= 0 {
		return types.DefaultParams().L2MinReporterRep * 1000
	}
	return params.L2MinReporterRep * 1000
}

func l2BudgetPerAgentMillis(params types.Params) int64 {
	return parseL2DecimalMillis(params.L2BudgetPerAgent, types.DefaultParams().L2BudgetPerAgent, DefaultL2BudgetPerAgentMillis)
}

func l2BudgetCapMillis(params types.Params) int64 {
	if params.L2BudgetCap <= 0 {
		return types.DefaultParams().L2BudgetCap * 1000
	}
	return params.L2BudgetCap * 1000
}

func l2NoEvidenceWeightMillis(params types.Params) int64 {
	return parseL2DecimalMillis(params.L2NoEvidenceWeight, types.DefaultParams().L2NoEvidenceWeight, 300)
}

func l2MutualReportPenaltyMillis(params types.Params) int64 {
	return parseL2DecimalMillis(params.L2MutualReportPenalty, types.DefaultParams().L2MutualReportPenalty, 100)
}

func l2AbuseThreshold(params types.Params) int64 {
	if params.L2AbuseThreshold <= 0 {
		return types.DefaultParams().L2AbuseThreshold
	}
	return params.L2AbuseThreshold
}

func parseL2DecimalMillis(raw string, fallback string, fallbackMillis int64) int64 {
	for _, candidate := range []string{raw, fallback} {
		if candidate == "" {
			continue
		}
		if d, err := math.LegacyNewDecFromStr(candidate); err == nil {
			return d.MulInt64(1000).TruncateInt64()
		}
	}
	return fallbackMillis
}
