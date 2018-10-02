package facts_test

import (
	"github.com/coinbase/dexter/facts"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestPlatformIsFact(t *testing.T) {
	assert := assert.New(t)

	check, ok := facts.Get("platform-is")
	assert.True(ok)

	assert.True(check.Assert([]string{runtime.GOOS}))
	assert.False(check.Assert([]string{"templeos"}))
}
