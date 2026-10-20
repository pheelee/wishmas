package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// @impl INT-001-AC1 INT-001-AC2 INT-001-AC4 INT-004-AC4 INT-004-AC5 CON-005
//
//go:embed messages.json
var messageData []byte

type message struct {
	EN string `json:"en"`
	DE string `json:"de"`
}

var messages = func() map[string]message {
	var result map[string]message
	if err := json.Unmarshal(messageData, &result); err != nil {
		panic(err)
	}
	return result
}()
var placeholders = regexp.MustCompile(`\{[a-z]+\}`)

func translate(locale, key string, args ...any) string {
	msg, ok := messages[key]
	if !ok {
		for _, candidate := range messages {
			if candidate.EN == key {
				msg = candidate
				ok = true
				break
			}
		}
	}
	if !ok {
		return key
	}
	text := msg.EN
	if locale == "de-CH" && msg.DE != "" {
		text = msg.DE
	}
	// Replace simultaneously: user content resembling a placeholder stays literal.
	names := placeholders.FindAllString(msg.EN, -1)
	values := map[string]string{}
	for i, name := range names {
		if i < len(args) {
			values[name] = fmt.Sprint(args[i])
		}
	}
	return placeholders.ReplaceAllStringFunc(text, func(name string) string {
		if value, ok := values[name]; ok {
			return value
		}
		return name
	})
}
func supportedLocale(s string) bool { return s == "en-US" || s == "de-CH" }
func householdLocale(s string) string {
	if s == "de-CH" {
		return s
	}
	return "en-US"
}

// @impl INT-002-AC1: respect browser preference order/quality, mapping all regional variants.
func browserLocale(header string) string {
	type preference struct {
		tag     string
		quality float64
	}
	var ordered []preference
	for _, entry := range strings.Split(header, ",") {
		parts := strings.Split(strings.TrimSpace(entry), ";")
		q := 1.0
		for _, parameter := range parts[1:] {
			if value, ok := strings.CutPrefix(strings.TrimSpace(parameter), "q="); ok {
				parsed, err := strconv.ParseFloat(value, 64)
				if err != nil || !(parsed >= 0 && parsed <= 1) {
					q = 0
				} else {
					q = parsed
				}
			}
		}
		ordered = append(ordered, preference{strings.ToLower(strings.TrimSpace(parts[0])), q})
	}
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].quality > ordered[j].quality })
	for _, p := range ordered {
		if p.quality <= 0 {
			continue
		}
		base, _, _ := strings.Cut(p.tag, "-")
		if base == "de" {
			return "de-CH"
		}
		if base == "en" {
			return "en-US"
		}
	}
	return "en-US"
}

// @impl INT-005-AC1 INT-005-AC2 INT-005-AC3: decimal-string formatting avoids floating-point loss.
func formatPrice(locale, raw string) string {
	if raw == "" {
		return ""
	}
	whole, fraction, _ := strings.Cut(raw, ".")
	whole = strings.TrimLeft(whole, "0")
	if whole == "" {
		whole = "0"
	}
	separator := ","
	if locale == "de-CH" {
		separator = "’"
	}
	var grouped strings.Builder
	for i := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			grouped.WriteString(separator)
		}
		grouped.WriteByte(whole[i])
	}
	return grouped.String() + "." + (fraction + "00")[:2]
}
