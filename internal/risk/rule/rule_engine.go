package rule

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/risk/model"
)

type RuleEngine struct {
	rules []*model.RiskRule
}

func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules: make([]*model.RiskRule, 0),
	}
}

func (re *RuleEngine) AddRule(rule *model.RiskRule) {
	re.rules = append(re.rules, rule)
	// Sort by priority
	for i := 0; i < len(re.rules)-1; i++ {
		for j := i + 1; j < len(re.rules); j++ {
			if re.rules[i].Priority > re.rules[j].Priority {
				re.rules[i], re.rules[j] = re.rules[j], re.rules[i]
			}
		}
	}
}

func (re *RuleEngine) RemoveRule(id string) {
	for i, r := range re.rules {
		if r.ID == id {
			re.rules = append(re.rules[:i], re.rules[i+1:]...)
			return
		}
	}
}

// Evaluate runs all enabled rules against the context and returns the first matching action.
func (re *RuleEngine) Evaluate(ctx context.Context, data map[string]interface{}) (*model.RiskEvent, error) {
	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}

		if re.matchRule(ctx, rule, data) {
			event := &model.RiskEvent{
				RuleName:  rule.Name,
				RiskLevel: rule.RiskLevel,
				RiskScore: rule.RiskScore,
				Action:    rule.Action,
				Reason:    fmt.Sprintf("matched rule: %s", rule.Name),
			}

			log.Ctx(ctx).Warn().
				Str("rule", rule.Name).
				Str("action", string(rule.Action)).
				Str("level", string(rule.RiskLevel)).
				Msg("risk rule matched")

			return event, nil
		}
	}

	// No rules matched - low risk
	return &model.RiskEvent{
		RiskLevel: model.RiskLow,
		RiskScore: 0,
		Action:    model.ActionAllow,
		Reason:    "no rules matched",
	}, nil
}

func (re *RuleEngine) matchRule(ctx context.Context, rule *model.RiskRule, data map[string]interface{}) bool {
	for _, cond := range rule.Conditions {
		if !re.evaluateCondition(cond, data) {
			return false
		}
	}
	return true // All conditions matched (AND logic)
}

func (re *RuleEngine) evaluateCondition(cond model.Condition, data map[string]interface{}) bool {
	val, ok := data[cond.Field]
	if !ok {
		return false
	}

	valStr := fmt.Sprintf("%v", val)

	switch cond.Operator {
	case "eq":
		return valStr == cond.Value
	case "ne":
		return valStr != cond.Value
	case "gt":
		return compareFloat(val, cond.Value) > 0
	case "lt":
		return compareFloat(val, cond.Value) < 0
	case "gte":
		return compareFloat(val, cond.Value) >= 0
	case "lte":
		return compareFloat(val, cond.Value) <= 0
	case "contains":
		return strings.Contains(strings.ToLower(valStr), strings.ToLower(cond.Value))
	case "in":
		values := strings.Split(cond.Value, ",")
		for _, v := range values {
			if strings.TrimSpace(v) == valStr {
				return true
			}
		}
		return false
	case "regex":
		// Simplified - exact match for now
		return valStr == cond.Value
	default:
		return false
	}
}

func compareFloat(value interface{}, target string) int {
	v := fmt.Sprintf("%v", value)
	if v == target {
		return 0
	}
	if v > target {
		return 1
	}
	return -1
}
