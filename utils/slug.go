package utils

import (
	"regexp"
	"strings"
	"time"
	"fmt"
	"math/rand"
)

func GenerateSlug(input string) string {

	slug := strings.ToLower(input)

	reg, _ := regexp.Compile("[^a-z0-9 ]+")
	slug = reg.ReplaceAllString(slug, "")

	slug = strings.ReplaceAll(slug, " ", "-")

	reg2, _ := regexp.Compile("-+")
	slug = reg2.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := r.Intn(9000) + 1000
	return fmt.Sprintf("%s-%d", slug, randomNumber)
}