package syncer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRandomPassword(t *testing.T) {
	data := make(map[string][]byte)
	addRandomPassword(data, "password")
	len := len(data["password"])
	t.Logf("password: %s", data["password"])
	assert.Equal(t, len, 24)
}
func TestGenerateASCIIPassword(t *testing.T) {
	password, _ := GenerateASCIIPassword(24)
	t.Logf("password: %s", password)
	assert.Equal(t, len(password), 24)
}
