package utils

import (
	"regexp"
	"strings"
)

func GenerateSlug(input string) string {

	slug := strings.ToLower(input)

	reg, _ := regexp.Compile("[^a-z0-9 ]+")
	slug = reg.ReplaceAllString(slug, "")

	slug = strings.ReplaceAll(slug, " ", "-")

	reg2, _ := regexp.Compile("-+")
	slug = reg2.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	return slug
}