package slack

import (
	"fmt"
	"strings"

	slackapi "github.com/slack-go/slack"

	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/slack/format"
	"github.com/autom8y/knossos/internal/trust"
)

// RenderResponse dispatches to the appropriate tier-specific renderer based on
// the response's confidence tier. Returns a slice of Slack Block Kit blocks.
func RenderResponse(resp *response.ReasoningResponse) []slackapi.Block {
	switch resp.Tier {
	case trust.TierHigh:
		return RenderHighConfidence(resp)
	case trust.TierMedium:
		return RenderMediumConfidence(resp)
	case trust.TierLow:
		return RenderLowConfidence(resp)
	default:
		return RenderLowConfidence(resp)
	}
}

// RenderHighConfidence renders a HIGH confidence response.
// Layout: answer text, divider, citation links, confidence indicator.
func RenderHighConfidence(resp *response.ReasoningResponse) []slackapi.Block {
	var blocks []slackapi.Block

	// Answer text — split into multiple blocks if > 3000 chars (Slack limit).
	blocks = append(blocks, splitAnswerBlocks(resp.Answer)...)

	// Divider before citations
	blocks = append(blocks, slackapi.NewDividerBlock())

	// Citation context blocks
	for _, citation := range resp.Citations {
		label := formatCitationLabel(citation)
		blocks = append(blocks, slackapi.NewContextBlock(
			"",
			slackapi.NewTextBlockObject(slackapi.MarkdownType, label, false, false),
		))
	}

	// Confidence indicator
	blocks = append(blocks, slackapi.NewContextBlock(
		"",
		slackapi.NewTextBlockObject(slackapi.MarkdownType,
			":large_green_circle: High confidence", false, false),
	))

	return blocks
}

// RenderMediumConfidence renders a MEDIUM confidence response.
// Layout: answer text, staleness warning, divider, citations, confidence indicator.
func RenderMediumConfidence(resp *response.ReasoningResponse) []slackapi.Block {
	var blocks []slackapi.Block

	// Answer text — split into multiple blocks if > 3000 chars (Slack limit).
	blocks = append(blocks, splitAnswerBlocks(resp.Answer)...)

	// Staleness warning with specific domains
	staleDomains := staleDomainList(resp)
	warningText := ":warning: *Some sources may not be current*"
	if staleDomains != "" {
		warningText += "\nStale domains: " + staleDomains
	}
	blocks = append(blocks, slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject(slackapi.MarkdownType, warningText, false, false),
		nil, nil,
	))

	// Divider before citations
	blocks = append(blocks, slackapi.NewDividerBlock())

	// Citation context blocks
	for _, citation := range resp.Citations {
		label := formatCitationLabel(citation)
		blocks = append(blocks, slackapi.NewContextBlock(
			"",
			slackapi.NewTextBlockObject(slackapi.MarkdownType, label, false, false),
		))
	}

	// Confidence indicator
	blocks = append(blocks, slackapi.NewContextBlock(
		"",
		slackapi.NewTextBlockObject(slackapi.MarkdownType,
			":large_yellow_circle: Medium confidence", false, false),
	))

	return blocks
}

// RenderLowConfidence renders a LOW confidence response (gap admission).
// Layout: header, gap reason, missing domains, stale domains, suggestions.
func RenderLowConfidence(resp *response.ReasoningResponse) []slackapi.Block {
	var blocks []slackapi.Block

	// Header
	blocks = append(blocks, slackapi.NewHeaderBlock(
		slackapi.NewTextBlockObject(slackapi.PlainTextType,
			":no_entry_sign: Cannot answer reliably", true, false),
	))

	// Gap reason
	reason := resp.Answer
	if resp.Gap != nil && resp.Gap.Reason != "" {
		reason = resp.Gap.Reason
	}
	blocks = append(blocks, slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject(slackapi.MarkdownType, reason, false, false),
		nil, nil,
	))

	// Missing domains section
	if resp.Gap != nil && len(resp.Gap.MissingDomains) > 0 {
		missingText := "*Missing knowledge domains:*\n"
		for _, d := range resp.Gap.MissingDomains {
			missingText += fmt.Sprintf("  - `%s`\n", d)
		}
		blocks = append(blocks, slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject(slackapi.MarkdownType, missingText, false, false),
			nil, nil,
		))
	}

	// Stale domains section
	if resp.Gap != nil && len(resp.Gap.StaleDomains) > 0 {
		staleText := "*Knowledge that may be outdated:*\n"
		for _, sd := range resp.Gap.StaleDomains {
			name := humanReadableName(sd.QualifiedName)
			if sd.DaysSinceGenerated > 0 {
				staleText += fmt.Sprintf("  - %s -- last updated %d days ago\n", name, sd.DaysSinceGenerated)
			} else {
				staleText += fmt.Sprintf("  - %s -- age unknown\n", name)
			}
		}
		blocks = append(blocks, slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject(slackapi.MarkdownType, staleText, false, false),
			nil, nil,
		))
	}

	// Suggestions
	if resp.Gap != nil && len(resp.Gap.Suggestions) > 0 {
		suggestText := "*Suggestions:*\n"
		for i, s := range resp.Gap.Suggestions {
			suggestText += fmt.Sprintf("%d. %s\n", i+1, s)
		}
		blocks = append(blocks, slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject(slackapi.MarkdownType, suggestText, false, false),
			nil, nil,
		))
	}

	return blocks
}

// formatCitationLabel formats a Citation as a human-readable Slack markdown string.
// Parses qualified names like "org::repo::domain" into "Domain (repo)".
func formatCitationLabel(c response.Citation) string {
	label := humanReadableName(c.QualifiedName)
	if c.Section != "" {
		label += fmt.Sprintf(" > %s", c.Section)
	}
	return fmt.Sprintf("*%s*", label)
}

// humanReadableName converts a qualified name "org::repo::domain" into
// a human-readable label like "Architecture (knossos)".
func humanReadableName(qualifiedName string) string {
	parts := strings.Split(qualifiedName, "::")
	if len(parts) != 3 {
		return qualifiedName
	}
	repo := parts[1]
	domain := parts[2]

	// Title-case the domain, replacing hyphens with spaces.
	display := strings.ReplaceAll(domain, "-", " ")
	display = simpleTitleCase(display)

	return fmt.Sprintf("%s (%s)", display, repo)
}

// staleDomainList returns a comma-separated list of stale domain names from the response.
func staleDomainList(resp *response.ReasoningResponse) string {
	if resp.Gap == nil || len(resp.Gap.StaleDomains) == 0 {
		// Check confidence score for stale info if gap is nil.
		return ""
	}
	names := make([]string, len(resp.Gap.StaleDomains))
	for i, sd := range resp.Gap.StaleDomains {
		names[i] = humanReadableName(sd.QualifiedName)
	}
	return strings.Join(names, ", ")
}

// simpleTitleCase capitalizes the first letter of each word, where word
// boundaries are spaces and slashes. Replaces deprecated strings.Title
// for simple ASCII domain names (e.g., "feat/materialization" -> "Feat/Materialization").
func simpleTitleCase(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	capitalizeNext := true
	for _, r := range s {
		if r == ' ' || r == '/' {
			b.WriteRune(r)
			capitalizeNext = true
		} else if capitalizeNext {
			b.WriteString(strings.ToUpper(string(r)))
			capitalizeNext = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// splitAnswerBlocks converts a GFM answer to Slack mrkdwn and splits it into
// multiple section blocks. Slack section block text has a 3000 character limit.
func splitAnswerBlocks(answer string) []slackapi.Block {
	const maxLen = 2900

	converted := format.Convert(answer)
	chunks := format.Chunk(converted, maxLen)

	blocks := make([]slackapi.Block, 0, len(chunks))
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		blocks = append(blocks, slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject(slackapi.MarkdownType, chunk, false, false),
			nil, nil,
		))
	}

	if len(blocks) == 0 {
		blocks = append(blocks, slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject(slackapi.MarkdownType, answer, false, false),
			nil, nil,
		))
	}

	return blocks
}

// RenderRateLimited returns blocks for a rate-limited response (TD-03).
func RenderRateLimited() []slackapi.Block {
	return []slackapi.Block{
		slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject(slackapi.MarkdownType,
				":hourglass: Clew is currently processing several requests. Please try again in a moment.",
				false, false),
			nil, nil,
		),
	}
}
