package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFromNetstringCorrect(t *testing.T) {
	input := "1:A,"
	expected := "A"
	result, err := FromNetstring(input)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, result)
}

func TestFromNetstringWrongLength(t *testing.T) {
	input := "2:A,"
	_, err := FromNetstring(input)
	if err == nil {
		t.Error("should throw wrong length error")
	}
}

