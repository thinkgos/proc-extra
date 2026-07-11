package sses

import (
	"testing"
)

func Test_Utils(t *testing.T) {
	t.Log(NextId())
	t.Log(NewEventId())
	t.Log(NewSessionId())
}
