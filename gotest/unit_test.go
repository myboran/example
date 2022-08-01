package gotest_test

import (
	"script/gotest"
	"testing"
)

func sub1(t *testing.T) {
	var a, b, expected = 1, 2, 3
	if gotest.Add(a, b) != expected {
		t.Errorf("Add(%d, %d) = %d; expected: %d", a, b, gotest.Add(a, b), expected)
	}
}

func sub2(t *testing.T) {
	var a, b, expected = 2, 2, 4
	if gotest.Add(a, b) != expected {
		t.Errorf("Add(%d, %d) = %d; expected: %d", a, b, gotest.Add(a, b), expected)
	}
}

func sub3(t *testing.T) {
	var a, b, expected = 1, 1, 2
	if gotest.Add(a, b) != expected {
		t.Errorf("Add(%d, %d) = %d; expected: %d", a, b, gotest.Add(a, b), expected)
	}
}

func TestSub(t *testing.T) {
	t.Run("A=1", sub1)
	t.Run("A=2", sub2)
	t.Run("b=1", sub3)
}
