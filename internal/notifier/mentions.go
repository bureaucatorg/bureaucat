package notifier

import (
	"regexp"

	"github.com/google/uuid"
)

// mentionPattern matches [@Display Name](/profile/<uuid>) markdown links
// produced by the MentionTextarea component.
var mentionPattern = regexp.MustCompile(`\[@[^\]]+\]\(/profile/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\)`)

// ParseMentions extracts unique user IDs from markdown mention links in the given text.
func ParseMentions(text string) []uuid.UUID {
	matches := mentionPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[uuid.UUID]struct{})
	var result []uuid.UUID
	for _, m := range matches {
		id, err := uuid.Parse(m[1])
		if err != nil {
			continue
		}
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

// DiffMentions returns user IDs that are in newText but not in oldText.
func DiffMentions(oldText, newText string) []uuid.UUID {
	oldMentions := ParseMentions(oldText)
	newMentions := ParseMentions(newText)

	oldSet := make(map[uuid.UUID]struct{}, len(oldMentions))
	for _, id := range oldMentions {
		oldSet[id] = struct{}{}
	}

	var added []uuid.UUID
	for _, id := range newMentions {
		if _, exists := oldSet[id]; !exists {
			added = append(added, id)
		}
	}
	return added
}
