package config

import (
	"strings"

	"github.com/google/uuid"
)

func generateString(prefix string) string {
	uuid := uuid.New().String()

	return prefix + strings.Replace(uuid, "-", "", -1)[:32-len(prefix)]
}

func generateRandomUser() string {
	return generateString("user")
}

func generateRandomGroup() string {
	return generateString("group")
}
