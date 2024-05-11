package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const scale = 2

var black color.RGBA = color.RGBA{75, 139, 190, 255}  //95,95,95
var white color.RGBA = color.RGBA{255, 232, 115, 255} //233,233,233

type POS struct {
	x int
	y int
}

func print_field(arr [][]int) {
	if len(arr) == 0 {
		return
	}

	loop(arr[0])
	print_field(arr[1:])
}

func loop(arr []int) {
	if len(arr) == 0 {
		fmt.Println()
		return
	}

	fmt.Print(arr[0], " ")
	loop(arr[1:])
}

func generate_row(size int) []int {
	var arr = []int{}
	return generate_row_rec(size, arr)
}

func generate_row_rec(size int, arr []int) []int {
	if len(arr) == size {
		return arr
	}

	if rand.Float32() < 0.5 {
		var new_arr = append(arr, 1)
		return generate_row_rec(size, new_arr)
	} else {
		var new_arr = append(arr, 0)
		return generate_row_rec(size, new_arr)
	}
}

func generate_field(size int) [][]int {
	var arr = [][]int{}
	return generate_field_rec(size, arr)
}

func generate_field_rec(size int, field [][]int) [][]int {
	if len(field) == size {
		return field
	}

	var row = generate_row(size)
	var new_field = append(field, row)

	return generate_field_rec(size, new_field)
}

// генерируем новое поколение
func gen_new_generation(size int, field [][]int) [][]int {
	var coord POS = POS{0, 0}
	var new_field = [][]int{}
	var new_generation = update_field(coord, size, field, new_field)
	return new_generation
}

// проход по элементам поля
func update_field(coord POS, size int, field [][]int, gen_field [][]int) [][]int {
	if len(gen_field) == size {
		return gen_field
	}

	var row = update_row(coord, size, field)
	var next_coord POS = POS{0, coord.y + 1}
	var new_field = append(gen_field, row)

	return update_field(next_coord, size, field, new_field)
}

func update_row(coord POS, size int, field [][]int) []int {
	var init_row = []int{}
	var new_row = update_row_rec(coord, size, field, init_row)
	return new_row
}

// обновляем статусы по строке и возвращаем ее полностью
func update_row_rec(coord POS, size int, field [][]int, new_row []int) []int {
	// смотрим выживет ли клетка в новом поколении
	var new_cell_status = get_next_cell_status(coord, size, field)

	// если строка закончилась, значит ничего не делаем
	if coord.x >= size {
		return new_row
	}

	var row = append(new_row, new_cell_status)
	var next_coord POS = POS{coord.x + 1, coord.y}
	return update_row_rec(next_coord, size, field, row)
}

// считаем число соседей для переданной клетки и определяем будет ли она живой
func get_next_cell_status(coord POS, size int, field [][]int) int {
	// проверяем все соседей
	var l_up = is_alive(coord.x-1, coord.y-1, size, field)
	var up = is_alive(coord.x, coord.y-1, size, field)
	var r_up = is_alive(coord.x+1, coord.y-1, size, field)
	var l = is_alive(coord.x-1, coord.y, size, field)
	var r = is_alive(coord.x+1, coord.y, size, field)
	var l_down = is_alive(coord.x-1, coord.y+1, size, field)
	var down = is_alive(coord.x, coord.y+1, size, field)
	var r_down = is_alive(coord.x+1, coord.y+1, size, field)

	// считаем число соседей
	var neigbours = l_up + up + r_up + l + r + l_down + down + r_down
	var is_i_alive = is_alive(coord.x, coord.y, size, field)

	// смотрим, что произойдет с клеткой в новом поколении
	// РОЖДЕНИЕ: если у пустой клетки есть 3 живых соседа, то она становится живой
	// ЭВОЛЮЦИЯ: если у живой клетки есть 2 или 3 живых соседа, то она не меняет свое состояние
	// СМЕРТЬ:   если у живой клетки меньше 2 или больше 3 живых соседей, то она умирает

	if neigbours == 3 {
		return 1
	} else if (neigbours == 3 || neigbours == 2) && is_i_alive == 1 {
		return field[coord.y][coord.x]
	}

	return 0 // клетка умирает, соседей либо > 2, либо < 3
}

// проверяем, живая ли ячейка
func is_alive(x int, y int, size int, field [][]int) int {
	// если вышли за поле, то там клетки нет
	if x > size-1 || x < 0 || y > size-1 || y < 0 {
		return 0
	}

	return field[y][x]
}

// Main
// func render(screen *ebiten.Image) {
// 	screen.Fill(white)
// 	for x := 0; x < WIDTH; x++ {
// 		for y := 0; y < HEIGHT; y++ {
// 			if grid[x][y] > 0 {
// 				for x1 := 0; x1 < scale; x1++ {
// 					for y1 := 0; y1 < scale; y1++ {
// 						screen.Set((x*scale)+x1, (y*scale)+y1, black)
// 					}
// 				}
// 			}
// 		}
// 	}
// }
// func frame(screen *ebiten.Image) error {
// 	counter++
// 	var err error = nil
// 	if counter == 20 {
// 		grid := gen_new_generation()
// 		err = gen_new_generation()
// 		counter = 0
// 	}
// 	if !ebiten.IsDrawingSkipped() {
// 		render(screen)
// 	}
// 	return err
// }

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *MyGame) Update() error {
	g.counter++
	if g.counter == 60 {
		g.field = gen_new_generation(g.size, g.field)
		g.counter = 0
	}
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *MyGame) Draw(screen *ebiten.Image) {
	// Write your game's rendering.
	screen.Fill(white)
	for x := 0; x < g.size; x++ {
		for y := 0; y < g.size; y++ {
			if g.field[x][y] == 1 {
				for x1 := 0; x1 < scale; x1++ {
					for y1 := 0; y1 < scale; y1++ {
						screen.Set((x*scale)+x1, (y*scale)+y1, black)
					}
				}
			}
		}
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *MyGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 640
}

func main() {
	// Specify the window size as you like. Here, a doubled size is specified.
	var _size = 640

	game := &MyGame{
		counter: 0,
		field:   generate_field(_size),
		size:    _size,
	}

	ebiten.SetWindowSize(game.size, game.size)
	ebiten.SetWindowTitle("Conway's game of life")

	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game implements ebiten.Game interface.
type MyGame struct {
	counter int
	field   [][]int
	size    int
}
