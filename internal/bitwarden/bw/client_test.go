package bw

import (
	"testing"

	test_command "github.com/maxlaverse/terraform-provider-bitwarden/internal/command/test"
	"github.com/stretchr/testify/assert"
)

func TestCreateObjectEncoding(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"encode":       `e30K`,
		"create  e30K": `{}`,
	})
	defer removeMocks(t)

	b := NewClient("dummy")
	_, err := b.CreateObject(Object{
		Type: ItemTypeLogin,
		Fields: []Field{
			{
				Name:  "test",
				Value: "passed",
				Type:  0,
			},
		},
	})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 2) {
		assert.Equal(t, "{\"login\":{},\"secureNote\":{},\"type\":1,\"fields\":[{\"name\":\"test\",\"value\":\"passed\",\"type\":0,\"linkedId\":null}]}:/:encode", commandsExecuted()[0])
		assert.Equal(t, "create  e30K", commandsExecuted()[1])
	}
}
