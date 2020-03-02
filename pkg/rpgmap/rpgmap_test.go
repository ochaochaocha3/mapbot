package rpgmap

import (
	"fmt"
	"testing"
)

func TestNewSquareMap(t *testing.T) {
	testcases := []struct {
		Width  int
		Height int
		Err    bool
	}{
		{Width: 10, Height: 10, Err: false},
		{Width: 20, Height: 10, Err: false},
		{Width: 0, Height: 1, Err: true},
		{Width: -1, Height: 1, Err: true},
		{Width: 1, Height: 0, Err: true},
		{Width: 1, Height: -1, Err: true},
		{Width: 0, Height: 0, Err: true},
		{Width: -1, Height: -1, Err: true},
	}

	for _, test := range testcases {
		name := fmt.Sprintf("%d x %d", test.Width, test.Height)
		t.Run(name, func(t *testing.T) {
			m, err := NewSquareMap(test.Width, test.Height)
			if err != nil {
				if test.Err {
					return
				}

				t.Fatalf("got err: %s", err)
			}

			if test.Err {
				t.Fatal("expected err")
			}

			actualWidth := m.Width()
			if actualWidth != test.Width {
				t.Errorf("Width: got %d, want %d", actualWidth, test.Width)
			}

			actualHeight := m.Height()
			if actualHeight != test.Height {
				t.Errorf("Height: got %d, want %d", actualHeight, test.Height)
			}
		})
	}
}

func TestSquareMap_AddChit(t *testing.T) {
	testcases := []struct {
		Width  int
		Height int
		Chit   Chit
		Err    bool
	}{
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				Name: "A",
				X:    0,
				Y:    0,
			},
			Err: false,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				Name: "A",
				X:    9,
				Y:    9,
			},
			Err: false,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				Name: "A",
				X:    -1,
				Y:    0,
			},
			Err: true,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				Name: "A",
				X:    0,
				Y:    -1,
			},
			Err: true,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				Name: "A",
				X:    10,
				Y:    9,
			},
			Err: true,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				Name: "A",
				X:    9,
				Y:    10,
			},
			Err: true,
		},
		{
			Width:  20,
			Height: 10,
			Chit: Chit{
				Name: "A",
				X:    10,
				Y:    9,
			},
			Err: false,
		},
		{
			Width:  10,
			Height: 20,
			Chit: Chit{
				Name: "A",
				X:    9,
				Y:    10,
			},
			Err: false,
		},
	}

	for _, test := range testcases {
		m, _ := NewSquareMap(test.Width, test.Height)
		name := fmt.Sprintf("%s << %s", m.SizeStr(), test.Chit.CoordStr())

		t.Run(name, func(t *testing.T) {
			err := m.AddChit(&test.Chit)

			if err != nil {
				if test.Err {
					return
				}

				t.Fatalf("got err: %s", err)
			}

			if test.Err {
				t.Fatal("expected err")
			}

			if _, ok := m.FindChit(test.Chit.Name); !ok {
				t.Fatal("added chit not found")
			}
		})
	}
}

func TestSquareMap_MoveChit(t *testing.T) {
	testcases := []struct {
		Width  int
		Height int
		Chit   Chit
		Err    bool
	}{
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				X: 0,
				Y: 0,
			},
			Err: false,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				X: 9,
				Y: 9,
			},
			Err: false,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				X: -1,
				Y: 0,
			},
			Err: true,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				X: 0,
				Y: -1,
			},
			Err: true,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				X: 10,
				Y: 9,
			},
			Err: true,
		},
		{
			Width:  10,
			Height: 10,
			Chit: Chit{
				X: 9,
				Y: 10,
			},
			Err: true,
		},
		{
			Width:  20,
			Height: 10,
			Chit: Chit{
				X: 10,
				Y: 9,
			},
			Err: false,
		},
		{
			Width:  10,
			Height: 20,
			Chit: Chit{
				X: 9,
				Y: 10,
			},
			Err: false,
		},
	}

	for _, test := range testcases {
		m, _ := NewSquareMap(test.Width, test.Height)
		m.AddChit(&Chit{Name: "A", X: 0, Y: 0})
		name := fmt.Sprintf("%s << %s", m.SizeStr(), test.Chit.CoordStr())

		t.Run(name, func(t *testing.T) {
			err := m.MoveChit("A", test.Chit.X, test.Chit.Y)

			if err != nil {
				if test.Err {
					return
				}

				t.Fatalf("got err: %s", err)
			}

			if test.Err {
				t.Fatal("expected err")
			}
		})
	}
}
