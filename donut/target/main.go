package main

import (
	"fmt"
	"math"
	"time"
)

const (
	ScreenWidth  = 80
	ScreenHeight = 22
	BufferLen    = ScreenWidth * ScreenHeight

	R1 = 1.0 // Torus tube radius (minor radius)
	R2 = 2.0 // Torus center offset (major radius)

	ZOffset = 5.0 // Camera fixed distance offset

	ScaleX = 30.0 // X projection scale factor
	ScaleY = 15.0 // Y projection scale factor

	ThetaSteps = 90
	PhiSteps   = 314

	ThetaSpacing = 2.0 * math.Pi / float64(ThetaSteps)
	PhiSpacing   = 2.0 * math.Pi / float64(PhiSteps)

	RotAInc = 0.30 // Rotation around X-axis increment
	RotBInc = 0.15 // Rotation around Z-axis increment

	LuminanceRamp = ".,-~:;=!*#$@"
)

// Point3D represents a point or vector in 3D space.
type Point3D struct {
	X, Y, Z float64
}

// Frame holds the data for one render tick.
type Frame struct {
	Pixels  [BufferLen]byte
	ZBuffer [BufferLen]float64
}

// Clear resets the buffers.
func (f *Frame) Clear() {
	for i := 0; i < BufferLen; i++ {
		f.Pixels[i] = ' '
		f.ZBuffer[i] = 0.0
	}
}

// Render prints the frame to the terminal output.
func (f *Frame) Render() {
	// A single output buffer avoids flickering and tearing
	out := []byte("\x1b[H") // Move cursor to top-left
	for y := 0; y < ScreenHeight; y++ {
		start := y * ScreenWidth
		end := start + ScreenWidth
		out = append(out, f.Pixels[start:end]...)
		out = append(out, '\n')
	}
	fmt.Print(string(out))
}

// rotatePoint applies rotation around X and Z axes (two-axis tumbling).
func rotatePoint(p Point3D, cosA, sinA, cosB, sinB float64) Point3D {
	// 1. Rotate around X axis by angle A
	y1 := p.Y*cosA - p.Z*sinA
	z1 := p.Y*sinA + p.Z*cosA

	// 2. Rotate around Z axis by angle B
	x2 := p.X*cosB - y1*sinB
	y2 := p.X*sinB + y1*cosB

	// Z unaffected by Z-axis rotation
	return Point3D{X: x2, Y: y2, Z: z1}
}

// projectToScreen applies perspective projection based on inverse depth.
func projectToScreen(p Point3D) (int, int, float64) {
	// Move away from camera
	zRot := p.Z + ZOffset
	ooz := 1.0 / zRot

	// Cast space onto 2D terminal plane
	screenX := int(40.0 + ScaleX*p.X*ooz)
	screenY := int(12.0 - ScaleY*p.Y*ooz)

	return screenX, screenY, ooz
}

// computeLighting aligns the normal to the light ray and finds the luminance.
func computeLighting(nx, ny, nz, cosA, sinA, cosB, sinB float64) int {
	normal := Point3D{nx, ny, nz}
	rotN := rotatePoint(normal, cosA, sinA, cosB, sinB)

	// Light direction L = (0, 1, -1) -> points up and into the screen
	L := rotN.Y*1.0 + rotN.Z*-1.0
	lumIndex := int(L * 8.0)

	// Handle boundaries for the luminance char map
	if lumIndex < 0 {
		lumIndex = 0
	} else if lumIndex >= len(LuminanceRamp) {
		lumIndex = len(LuminanceRamp) - 1
	}

	return lumIndex
}

// ComputeTorus evaluates all surface points for rendering.
func (f *Frame) ComputeTorus(A, B float64) {
	cosA, sinA := math.Cos(A), math.Sin(A)
	cosB, sinB := math.Cos(B), math.Sin(B)

	for thetaIdx := 0; thetaIdx < ThetaSteps; thetaIdx++ {
		theta := float64(thetaIdx) * ThetaSpacing
		cosTheta, sinTheta := math.Cos(theta), math.Sin(theta)

		for phiIdx := 0; phiIdx < PhiSteps; phiIdx++ {
			phi := float64(phiIdx) * PhiSpacing
			cosPhi, sinPhi := math.Cos(phi), math.Sin(phi)

			// Calculate local coordinates of the 3D Torus
			// (R2 + R1*cosPhi) places the tube off-center
			r := R2 + R1*cosPhi
			x := r * cosTheta
			y := r * sinTheta
			z := R1 * sinPhi

			// Surface normals (tube vectors unaffected by major radius R2)
			nx := cosPhi * cosTheta
			ny := cosPhi * sinTheta
			nz := sinPhi

			// Apply global rotations for this frame
			pRotated := rotatePoint(Point3D{x, y, z}, cosA, sinA, cosB, sinB)

			// Map back to screen perspective
			screenX, screenY, ooz := projectToScreen(pRotated)

			// Only process points that land in our visible screen plane
			if screenX >= 0 && screenX < ScreenWidth && screenY >= 0 && screenY < ScreenHeight {
				
				// Standard Z-buffer check to occlude obscured points
				idx := screenY*ScreenWidth + screenX
				if ooz > f.ZBuffer[idx] {
					f.ZBuffer[idx] = ooz

					lumIdx := computeLighting(nx, ny, nz, cosA, sinA, cosB, sinB)
					f.Pixels[idx] = LuminanceRamp[lumIdx]
				}
			}
		}
	}
}

func main() {
	var frame Frame

	fmt.Print("\x1b[2J") // Clear terminal fully before start

	var A, B float64
	for {
		frame.Clear()
		frame.ComputeTorus(A, B)
		frame.Render()

		A += RotAInc
		B += RotBInc

		// Throttling prevents uncontrolled console spam
		time.Sleep(40 * time.Millisecond)
	}
}
