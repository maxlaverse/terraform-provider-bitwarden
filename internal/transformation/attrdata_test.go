package transformation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeSet struct {
	items []interface{}
}

func (s fakeSet) List() []interface{} { return s.items }

func TestAsInterfaceList(t *testing.T) {
	t.Parallel()

	list, ok := asInterfaceList([]interface{}{"a", "b"})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"a", "b"}, list)

	list, ok = asInterfaceList([]string{"x", "y"})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"x", "y"}, list)

	list, ok = asInterfaceList(fakeSet{items: []interface{}{"1"}})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"1"}, list)

	_, ok = asInterfaceList(42)
	assert.False(t, ok)
}
