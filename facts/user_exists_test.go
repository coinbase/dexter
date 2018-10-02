package facts_test

import (
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/facts"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserExistsFact(t *testing.T) {
	assert := assert.New(t)

	helpers.StubLocalUsers([]string{"root", "foo"})
	check, ok := facts.Get("user-exists")
	assert.True(ok)

	salt := "foobar01"
	check.Salt = salt

	assert.True(check.Assert([]string{facts.Hash("root", salt)}))
	assert.False(check.Assert([]string{facts.Hash("bar", salt)}))
}
