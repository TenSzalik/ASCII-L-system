/*
TODO:

- settings file
- enhance performance (remove grid?)
- validators
- support for small signs like 'f'
- support colors for leafs
- support changing bg color
*/

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
)

// Compose axiom into long chain defined by rules - here we call it composition.
//
// For example:
// axiom="X", rules="X=FX;F=FF"
// 1 iteration:
//     X -> FX
// 2 iter:
//     FX -> FFFX
// 3 iter:
//     FFFX -> FFFFFFFX
// etc.
func seedComposer(rules, axiom string, iterations int) string {
	// map rules
	rulesList := strings.Split(rules, ";")
	rulesMap := make(map[string]string)
	for _, rule := range rulesList {
		parts := strings.SplitN(rule, "=", 2)
		if len(parts) == 2 {
			rulesMap[parts[0]] = parts[1]
		}
	}

	// generate chain defined by rules 
	composition := axiom
	for i := 0; i < iterations; i++ {
		var sb strings.Builder
		for _, ch := range composition {
			if replacement, ok := rulesMap[string(ch)]; ok {
				sb.WriteString(replacement)
			} else {
				sb.WriteRune(ch)
			}
		}
		composition = sb.String()
	}
	return composition
}

type Grid struct {
	width  int
	height int
	data   [][]rune
}

// Create grid. We need grid to store our plant.
// This approach is handy, but so slow... We need to consider
// something faster.
func NewGrid(width, height int) *Grid {
	grid := &Grid{
		width:  width,
		height: height,
		data:   make([][]rune, height),
	}

	for i := range grid.data {
		grid.data[i] = make([]rune, width)
		for j := range grid.data[i] {
			grid.data[i][j] = ' '
		}
	}
	return grid
}

// Display ASCII in terminal
func (g *Grid) GenerateAscii() {
	for _, row := range g.data {
		fmt.Println(string(row))
	}
}

// Save picture
func (g *Grid) SaveImage(filename string) {
	img := image.NewRGBA(image.Rect(0, 0, g.width, g.height))
	white := color.RGBA{255, 170, 0, 255}
	black := color.RGBA{16, 16, 16, 255}

	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			if g.data[y][x] != ' ' {
				img.Set(x, y, white)
			} else {
				img.Set(x, y, black)
			}
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Something went wrong:", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Println("Error saving image:", err)
	} else {
		fmt.Println("Image saved to", filename)
	}
}

type State struct {
	x, y, angle float64
}

// Grow plant
func (g *Grid) seedGrower(instructions string, startX, startY, startAngle int, angleDelta float64) {
	state := State{x: float64(startX), y: float64(startY), angle: float64(startAngle)}
	var stack []State

	for _, ch := range instructions {
		switch ch {
		case 'F':
			rad := state.angle * math.Pi / 180.0
			newX := state.x + math.Cos(rad)
			newY := state.y - math.Sin(rad)
			x := int(math.Round(newX))
			y := int(math.Round(newY))

			// Check if coordinates covers grid.
			// Temporary solution.
			// This code is so bad, we don't respect that and it has to be changed.
			if x >= 0 && x < g.width && y >= 0 && y < g.height {
				g.data[y][x] = '*'
			}
			state.x = newX
			state.y = newY
		case '+':
			state.angle += angleDelta
		case '-':
			state.angle -= angleDelta
		case '[':
			stack = append(stack, state)
		case ']':
			if len(stack) > 0 {
				state = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
		}
	}
}

func main() {
	var _axiom string
	var _rules string
	var _angle float64
	var _iterations int
	var _rows int
	var _cols int
	var _outputFile string
	var _stream string
	var _start string
	var _startAngle int

	flag.StringVar(&_axiom, "axiom", "F", "Initial axiom of the L-system")
	flag.StringVar(&_rules, "rules", "F=F", "L-system rules in the format 'F=FF;X=FX'")
	flag.Float64Var(&_angle, "angle", 25.0, "Rotation angle for '+' and '-' in degrees")
	flag.IntVar(&_iterations, "iterations", 4, "Number of L-system iterations")
	flag.IntVar(&_rows, "rows", 80, "Grid height")
	flag.IntVar(&_cols, "cols", 80, "Grid width")
	flag.StringVar(&_outputFile, "output", "output.png", "Output file name (e.g., output.png)")
	flag.StringVar(&_stream, "stream", "both", "Kind of stream: ascii, img, both")
	flag.StringVar(&_start, "start", "bottom", "Start drawing from: bottom, middle, top")
	flag.IntVar(&_startAngle, "startangle", 90, "Start drawing from angle: 0, 90, 180, 270") // draw upwards
	flag.Parse()

	var startX int
	var startY int

	switch _start {
	case "bottom":
		startX = _cols / 2
		startY = _rows - 1
	case "middle":
		startX = _cols / 2
		startY = _rows / 2
	case "top":
		startX = _cols / 2
		startY = 1
	}

	composition := seedComposer(_rules, _axiom, _iterations)
	grid := NewGrid(_cols, _rows)
	grid.seedGrower(composition, startX, startY, _startAngle, _angle)

	switch _stream {
	case "both":
		grid.GenerateAscii()
		grid.SaveImage(_outputFile)
	case "ascii":
		grid.GenerateAscii()
	case "img":
		grid.SaveImage(_outputFile)
	}
}
