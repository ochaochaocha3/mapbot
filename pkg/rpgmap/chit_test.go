package rpgmap

import (
	"testing"
)

func TestChit_CoordStr(t *testing.T) {
	testcases := []struct {
		X        int
		Y        int
		expected string
	}{
		{X: 1, Y: 2, expected: "(2, 3)"},
		{X: 0, Y: 1, expected: "(1, 2)"},
	}

	for _, test := range testcases {
		t.Run(test.expected, func(t *testing.T) {
			c := Chit{
				X: test.X,
				Y: test.Y,
			}

			actual := c.CoordStr()
			if actual != test.expected {
				t.Fatalf("got: %s, want: %s", actual, test.expected)
			}
		})
	}
}

func TestChit_String(t *testing.T) {
	testcases := []struct {
		Name     string
		X        int
		Y        int
		expected string
	}{
		{Name: "ゆうしゃ", X: 1, Y: 2, expected: "ゆうしゃ (2, 3)"},
		{Name: "Bob", X: 0, Y: 1, expected: "Bob (1, 2)"},
	}

	for _, test := range testcases {
		t.Run(test.expected, func(t *testing.T) {
			c := Chit{
				Name: test.Name,
				X:    test.X,
				Y:    test.Y,
			}

			actual := c.String()
			if actual != test.expected {
				t.Fatalf("got: %s, want: %s", actual, test.expected)
			}
		})
	}
}
