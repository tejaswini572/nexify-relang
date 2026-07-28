package main

// Direction enum
type Direction int

const (
	UP Direction = iota
	RIGHT
	DOWN
	LEFT
)

// PipeStyle is mapped directly via index (0-9)

// Pipe tracks the state of a single pipe
type Pipe struct {
	X         int
	Y         int
	Direction Direction
	PipeType  int
	Color     int
}
