package dao

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRelationAction(t *testing.T) {
	err := RelationAction(12182603931062272, 10760536648060928, 1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = RelationAction(12182603931062272, 10760536648060928, -1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestRelationActionErr(t *testing.T) {
	err := RelationAction(12182603931062272, 10760536648060928, 1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = RelationAction(12182603931062272, 10760536648060928, 1)
	assert.Equal(t, ErrAlreadyFollow, err)
	err = RelationAction(12182603931062272, 10760536648060928, -1)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	time.Sleep(delayTime)
	err = RelationAction(12182603931062272, 10760536648060928, -1)
	assert.Equal(t, ErrNotFollow, err)
}
