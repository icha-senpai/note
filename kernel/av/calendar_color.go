// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package av

import (
	"errors"
	"regexp"
)

// Calendar uses the table's ordered fields and values with a separate presentation.
type CalendarConfig struct {
	DateKeyID string `json:"dateKeyID"`
}
type ConditionalColorRule struct {
	ID      string      `json:"id"`
	Enabled bool        `json:"enabled"`
	Color   string      `json:"color"`
	Filter  *ViewFilter `json:"filter"`
}

var conditionalColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var selectColorPattern = regexp.MustCompile(`^(?:[1-9]|1[0-4])$`)

// Empty colors occur in inherited options; new custom colors use opaque HEX.
func ValidSelectColor(color string) bool {
	return color == "" || selectColorPattern.MatchString(color) || conditionalColorPattern.MatchString(color)
}

func ValidateConditionalColors(view *AttributeView, rules []*ConditionalColorRule) error {
	if len(rules) > 50 {
		return errors.New("too many color rules")
	}
	seen := map[string]bool{}
	var validate func(*ViewFilter) error
	validate = func(filter *ViewFilter) error {
		if filter == nil {
			return errors.New("missing color condition")
		}
		if filter.IsGroup() {
			if len(filter.Filters) == 0 || (filter.Combination != FilterCombinationAnd && filter.Combination != FilterCombinationOr) {
				return errors.New("invalid color condition group")
			}
			for _, child := range filter.Filters {
				if err := validate(child); err != nil {
					return err
				}
			}
			return nil
		}
		key, err := view.GetKey(filter.Column)
		if err != nil {
			return err
		}
		if filter.Value == nil || filter.Value.Type != key.Type {
			return errors.New("invalid color condition value")
		}
		switch key.Type {
		case KeyTypeBlock, KeyTypeText, KeyTypeNumber, KeyTypeDate, KeyTypeCreated, KeyTypeUpdated,
			KeyTypeSelect, KeyTypeMSelect, KeyTypeCheckbox, KeyTypeURL, KeyTypeEmail, KeyTypePhone:
		default:
			return errors.New("unsupported color condition field")
		}
		switch filter.Operator {
		case FilterOperatorIsEqual, FilterOperatorIsNotEqual, FilterOperatorIsGreater, FilterOperatorIsGreaterOrEqual,
			FilterOperatorIsLess, FilterOperatorIsLessOrEqual, FilterOperatorContains, FilterOperatorDoesNotContain,
			FilterOperatorIsEmpty, FilterOperatorIsNotEmpty, FilterOperatorStartsWith, FilterOperatorEndsWith,
			FilterOperatorIsBetween, FilterOperatorIsTrue, FilterOperatorIsFalse:
		default:
			return errors.New("invalid color condition operator")
		}
		return nil
	}
	for _, rule := range rules {
		if rule == nil || rule.ID == "" || seen[rule.ID] || !conditionalColorPattern.MatchString(rule.Color) {
			return errors.New("invalid color rule")
		}
		seen[rule.ID] = true
		if err := ValidateFilterDepth([]*ViewFilter{rule.Filter}); err != nil {
			return err
		}
		if err := validate(rule.Filter); err != nil {
			return err
		}
	}
	return nil
}

// Color conditions use the same values and evaluator as ordinary database filters.
// First matching rule wins; unavailable fields never color entries accidentally.
func ConditionalItemColors(viewable Viewable, attrView *AttributeView, rules []*ConditionalColorRule,
	rollups map[string]Collection, cached map[string]*AttributeView) map[string]string {
	colors := map[string]string{}
	collection := viewable.(Collection)
	for _, rule := range rules {
		if rule == nil || !rule.Enabled || !conditionalColorPattern.MatchString(rule.Color) {
			continue
		}
		if ValidateConditionalColors(attrView, []*ConditionalColorRule{rule}) != nil {
			continue
		}
		indexes := map[string]int{}
		collectLeafColumnIndexes([]*ViewFilter{rule.Filter}, collection.GetFields(), indexes)
		for _, item := range collection.GetItems() {
			if _, matched := colors[item.GetID()]; matched {
				continue
			}
			if evalNode(rule.Filter, item.GetValues(), indexes, attrView, item.GetID(), rollups, cached) {
				colors[item.GetID()] = rule.Color
			}
		}
	}
	return colors
}
