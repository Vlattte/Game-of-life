package main

import (
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const scale = 4

var black color.RGBA = color.RGBA{75, 139, 190, 255}  //95,95,95
var white color.RGBA = color.RGBA{255, 232, 115, 255} //233,233,233

const (
	screenWidth  = 740
	screenHeight = 640

	gameWidth  = 640 / 4
	gameHeight = 640 / 4
)

type POS struct {
	x int
	y int
}

func generate_row(size int) []byte {
	var arr = []byte{}
	return generate_row_rec(size, arr)
}

func generate_row_rec(size int, arr []byte) []byte {
	if len(arr) == size {
		return arr
	}

	if rand.Float32() < 0.1 {
		var new_arr = append(arr, 1)
		return generate_row_rec(size, new_arr)
	} else {
		var new_arr = append(arr, 0)
		return generate_row_rec(size, new_arr)
	}
}

func generate_field(size int) [][]byte {
	var arr = [][]byte{}
	return generate_field_rec(size, arr)
}

func generate_field_rec(size int, field [][]byte) [][]byte {
	if len(field) == size {
		return field
	}

	var row = generate_row(size)
	var new_field = append(field, row)

	return generate_field_rec(size, new_field)
}

// генерируем новое поколение
func gen_new_generation(size int, field [][]byte) [][]byte {
	var coord POS = POS{0, 0}
	var new_field = [][]byte{}
	var new_generation = update_field(coord, size, field, new_field)
	return new_generation
}

// проход по элементам поля
func update_field(coord POS, size int, field [][]byte, gen_field [][]byte) [][]byte {
	if len(gen_field) == size {
		return gen_field
	}

	var row = update_row(coord, size, field)
	var next_coord POS = POS{0, coord.y + 1}
	var new_field = append(gen_field, row)

	return update_field(next_coord, size, field, new_field)
}

func update_row(coord POS, size int, field [][]byte) []byte {
	var init_row = []byte{}
	var new_row = update_row_rec(coord, size, field, init_row)
	return new_row
}

// обновляем статусы по строке и возвращаем ее полностью
func update_row_rec(coord POS, size int, field [][]byte, new_row []byte) []byte {
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
func get_next_cell_status(coord POS, size int, field [][]byte) byte {
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
func is_alive(x int, y int, size int, field [][]byte) byte {
	// если вышли за поле, то там клетки нет
	if x > size-1 || x < 0 || y > size-1 || y < 0 {
		return 0
	}

	return field[y][x]
}

type touch struct {
	id  ebiten.TouchID
	pos pos
}

type pos struct {
	x int
	y int
}

// Game implements ebiten.Game interface.
type MyGame struct {
	// состояние игры
	counter  int
	is_pause bool

	field  [][]byte
	width  int
	height int

	// сдвиги координат, чтобы массив с полем был больше отображаемой области
	x_offset int
	y_offset int

	cursor pos

	canvasImage *ebiten.Image
}

func NewGame(maxInitLiveCells int) *MyGame {
	g := &MyGame{
		canvasImage: ebiten.NewImage(screenWidth, screenHeight),
		counter:     10,
		is_pause:    false,
		width:       gameWidth,
		height:      gameHeight,
		x_offset:    gameWidth / 2,
		y_offset:    gameHeight / 2,
	}

	g.init(maxInitLiveCells)

	return g
}

// init inits MyGame with a random state.
func (g *MyGame) init(maxLiveCells int) {
	g.field = make([][]byte, g.width, g.height)
	for i := 0; i < g.height; i++ {
		for j := 0; j < g.width; j++ {
			g.field[i] = append(g.field[i], 0)
		}
	}

	for i := 0; i < maxLiveCells; i++ {
		x := rand.Intn(g.height)
		y := rand.Intn(g.width)
		g.field[x][y] = 1
	}
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *MyGame) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.is_pause = !g.is_pause
	}

	// если пауза, то не обновляем игру
	if !g.is_pause {
		g.counter++
	}

	// переходим к следующему поколению
	if g.counter == 20 {
		g.field = gen_new_generation(g.width, g.field)
		g.counter = 0
	}

	// рисуем пиксели, если нарисовали в игровой зоне
	mx, my := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.paint(g.canvasImage, mx, my)
	}
	g.cursor = pos{
		x: mx,
		y: my,
	}

	return nil
}

// paint draws the brush on the given canvas image at the position (x, y).
func (g *MyGame) paint(canvas *ebiten.Image, x, y int) {
	if x < gameWidth*4 && y < gameWidth*4 {
		g.field[x/4][y/4] = 1
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *MyGame) Draw(screen *ebiten.Image) {
	// очищаем экран
	screen.Fill(white)
	// screen.DrawImage(g.canvasImage, nil)

	// рисуем линии отделяющие шаблонные фигуры

	// вывод положения курсора
	// msg := fmt.Sprintf("(%d, %d)", g.cursor.x, g.cursor.y)
	// ebitenutil.DebugPrint(screen, msg)

	for x := 0; x < gameHeight; x++ {
		for y := 0; y < gameWidth; y++ {
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
func (g *MyGame) Layout(outsideWidth, outsideHeight int) (_screenWidth, _screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Conway's game of life")

	// Call ebiten.RunGame to start your game loop.
	// int((screenWidth * screenHeight) / 100)
	if err := ebiten.RunGame(NewGame(int((gameWidth * gameWidth) / 10))); err != nil {
		log.Fatal(err)
	}
}
