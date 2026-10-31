package database

import "strings"

// LikeContainsPattern escapes LIKE metacharacters and adds a substring wildcard.
func LikeContainsPattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(value) + "%"
}
