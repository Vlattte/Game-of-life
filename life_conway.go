package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const scale = 4

var black color.RGBA = color.RGBA{75, 139, 190, 255}  //95,95,95
var white color.RGBA = color.RGBA{255, 232, 115, 255} //233,233,233

const (
	// реальный размер отображаемой области в пикселях
	screenWidth  = 700
	screenHeight = 640

	// делим на scale, потому что каждая клетка массива field занимаем scale на scale пикселей
	gameWidth  = 640 / scale
	gameHeight = 640 / scale
)

type POS struct {
	x int
	y int
}

func generate_row(width int) []byte {
	var arr = []byte{}
	return generate_row_rec(width, arr)
}

func generate_row_rec(width int, arr []byte) []byte {
	if len(arr) == width {
		return arr
	}

	if rand.Float32() < 0.1 {
		var new_arr = append(arr, 1)
		return generate_row_rec(width, new_arr)
	} else {
		var new_arr = append(arr, 0)
		return generate_row_rec(width, new_arr)
	}
}

func generate_field(height int, width int) [][]byte {
	var arr = [][]byte{}
	return generate_field_rec(height, width, arr)
}

func generate_field_rec(height int, width int, field [][]byte) [][]byte {
	if len(field) == height {
		return field
	}

	var row = generate_row(width)
	var new_field = append(field, row)

	return generate_field_rec(height, width, new_field)
}

// генерируем новое поколение
func gen_new_generation(height int, width int, field [][]byte) [][]byte {
	var coord POS = POS{0, 0}
	var new_field = [][]byte{}
	var new_generation = update_field(coord, height, width, field, new_field)
	return new_generation
}

// проход по элементам поля
func update_field(coord POS, height int, width int, field [][]byte, gen_field [][]byte) [][]byte {
	if len(gen_field) == height {
		return gen_field
	}

	var row = update_row(coord, height, width, field)
	var next_coord POS = POS{coord.x + 1, 0}
	var new_field = append(gen_field, row)

	return update_field(next_coord, height, width, field, new_field)
}

func update_row(coord POS, height int, width int, field [][]byte) []byte {
	var init_row = []byte{}
	var new_row = update_row_rec(coord, height, width, field, init_row)
	return new_row
}

// обновляем статусы по строке и возвращаем ее полностью
func update_row_rec(coord POS, height int, width int, field [][]byte, new_row []byte) []byte {
	// смотрим выживет ли клетка в новом поколении
	var new_cell_status = get_next_cell_status(coord, height, width, field)

	// если строка закончилась, значит ничего не делаем
	if coord.y >= width {
		return new_row
	}

	var row = append(new_row, new_cell_status)
	var next_coord POS = POS{coord.x, coord.y + 1}
	return update_row_rec(next_coord, height, width, field, row)
}

// считаем число соседей для переданной клетки и определяем будет ли она живой
func get_next_cell_status(coord POS, height int, width int, field [][]byte) byte {
	// проверяем все соседей
	var l_up = is_alive(coord.x-1, coord.y-1, height, width, field)
	var up = is_alive(coord.x, coord.y-1, height, width, field)
	var r_up = is_alive(coord.x+1, coord.y-1, height, width, field)
	var l = is_alive(coord.x-1, coord.y, height, width, field)
	var r = is_alive(coord.x+1, coord.y, height, width, field)
	var l_down = is_alive(coord.x-1, coord.y+1, height, width, field)
	var down = is_alive(coord.x, coord.y+1, height, width, field)
	var r_down = is_alive(coord.x+1, coord.y+1, height, width, field)

	// считаем число соседей
	var neigbours = l_up + up + r_up + l + r + l_down + down + r_down
	var is_i_alive = is_alive(coord.x, coord.y, height, width, field)

	// смотрим, что произойдет с клеткой в новом поколении
	// РОЖДЕНИЕ: если у пустой клетки есть 3 живых соседа, то она становится живой
	// ЭВОЛЮЦИЯ: если у живой клетки есть 2 или 3 живых соседа, то она не меняет свое состояние
	// СМЕРТЬ:   если у живой клетки меньше 2 или больше 3 живых соседей, то она умирает

	if neigbours == 3 {
		return 1
	} else if (neigbours == 3 || neigbours == 2) && is_i_alive == 1 {
		return field[coord.x][coord.y]
	}

	return 0 // клетка умирает, соседей либо > 2, либо < 3
}

// проверяем, живая ли ячейка
func is_alive(x int, y int, height int, width int, field [][]byte) byte {
	// если вышли за поле, то там клетки нет
	if x > height-1 || x < 0 || y > width-1 || y < 0 {
		return 0
	}

	return field[x][y]
}

// Game implements ebiten.Game interface.
type MyGame struct {
	// состояние игры
	counter     int
	max_counter int

	is_pause       bool
	is_figure_draw bool

	field  [][]byte
	pixels []PIXEL

	// реальный размер всего поля == размер массива field с учетом расширения
	// постоянно меняется, по мере движения пикселей
	width  int
	height int

	// сдвиги координат, чтобы массив с полем был больше отображаемой области
	// например, field размером 1000 на 1000, тогда x_offset = 250, y_offset = 250
	// так будем показывать пиксели начиная с 250 по x и 250 по y
	x_offset int
	y_offset int

	cursor POS

	ui *ebitenui.UI
	// btn *widget.Button
}

func NewGame(maxInitLiveCells int) *MyGame {
	g := &MyGame{
		counter:        10,
		max_counter:    20,
		is_pause:       true,
		is_figure_draw: false,
		width:          gameWidth,
		height:         gameHeight,
		x_offset:       0,
		y_offset:       0,
		// btn:      button,
	}

	// load images for button states: idle, hover, and pressed

	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10))),
	)

	// Gosper Glider Gun
	block_image, _ := loadButtonImage("patterns/Gosper_Glider_Gun.png")
	button_block := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			}),
		),

		// specify the images to use
		widget.ButtonOpts.Image(block_image),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// Gosper Glider Gun
			clear(g.pixels)
			g.is_figure_draw = true
			g.pixels = append(g.pixels, PIXEL{0, 0, 1})
			g.pixels = append(g.pixels, PIXEL{-1, 0, 1})
			g.pixels = append(g.pixels, PIXEL{-1, 1, 1})
			g.pixels = append(g.pixels, PIXEL{-2, 2, 1})
			g.pixels = append(g.pixels, PIXEL{-1, -1, 1})
			g.pixels = append(g.pixels, PIXEL{-2, -2, 1})
			g.pixels = append(g.pixels, PIXEL{-3, 0, 1})

			g.pixels = append(g.pixels, PIXEL{-4, -3, 1})
			g.pixels = append(g.pixels, PIXEL{-4, 3, 1})
			g.pixels = append(g.pixels, PIXEL{-5, -3, 1})
			g.pixels = append(g.pixels, PIXEL{-5, 3, 1})
			g.pixels = append(g.pixels, PIXEL{-6, -2, 1})
			g.pixels = append(g.pixels, PIXEL{-6, 2, 1})
			g.pixels = append(g.pixels, PIXEL{-7, -1, 1})
			g.pixels = append(g.pixels, PIXEL{-7, 1, 1})
			g.pixels = append(g.pixels, PIXEL{-7, 0, 1})

			g.pixels = append(g.pixels, PIXEL{-16, 0, 1})
			g.pixels = append(g.pixels, PIXEL{-16, -1, 1})
			g.pixels = append(g.pixels, PIXEL{-17, 0, 1})
			g.pixels = append(g.pixels, PIXEL{-17, -1, 1})

			g.pixels = append(g.pixels, PIXEL{3, -1, 1})
			g.pixels = append(g.pixels, PIXEL{3, -2, 1})
			g.pixels = append(g.pixels, PIXEL{3, -3, 1})
			g.pixels = append(g.pixels, PIXEL{4, -1, 1})
			g.pixels = append(g.pixels, PIXEL{4, -2, 1})
			g.pixels = append(g.pixels, PIXEL{4, -3, 1})
			g.pixels = append(g.pixels, PIXEL{5, -4, 1})
			g.pixels = append(g.pixels, PIXEL{5, 0, 1})
			g.pixels = append(g.pixels, PIXEL{7, -4, 1})
			g.pixels = append(g.pixels, PIXEL{7, 0, 1})
			g.pixels = append(g.pixels, PIXEL{7, -5, 1})
			g.pixels = append(g.pixels, PIXEL{7, 1, 1})

			g.pixels = append(g.pixels, PIXEL{17, -2, 1})
			g.pixels = append(g.pixels, PIXEL{17, -3, 1})
			g.pixels = append(g.pixels, PIXEL{18, -2, 1})
			g.pixels = append(g.pixels, PIXEL{18, -3, 1})
		}),
	)

	// GLIDER
	glider_image, _ := loadButtonImage("patterns/glider.png")
	button_glider := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			}),
		),

		// specify the images to use
		widget.ButtonOpts.Image(glider_image),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// GLIDER
			clear(g.pixels)
			g.is_figure_draw = true
			g.pixels = append(g.pixels, PIXEL{0, 0, 1})
			g.pixels = append(g.pixels, PIXEL{1, 1, 1})
			g.pixels = append(g.pixels, PIXEL{2, 0, 1})
			g.pixels = append(g.pixels, PIXEL{2, -1, 1})
			g.pixels = append(g.pixels, PIXEL{2, 1, 1})
		}),
	)

	// PULSAR
	pulsar_image, _ := loadButtonImage("patterns/pulsar.png")
	button_pulsar := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			}),
		),

		// specify the images to use
		widget.ButtonOpts.Image(pulsar_image),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// GLIDER
			clear(g.pixels)
			g.is_figure_draw = true
			g.pixels = append(g.pixels, PIXEL{0, 0, 1})
			g.pixels = append(g.pixels, PIXEL{1, 0, 1})
			g.pixels = append(g.pixels, PIXEL{2, 0, 1})
			g.pixels = append(g.pixels, PIXEL{2, 1, 1})

			g.pixels = append(g.pixels, PIXEL{0, 6, 1})
			g.pixels = append(g.pixels, PIXEL{1, 6, 1})
			g.pixels = append(g.pixels, PIXEL{2, 6, 1})
			g.pixels = append(g.pixels, PIXEL{2, 5, 1})

			g.pixels = append(g.pixels, PIXEL{4, -2, 1})
			g.pixels = append(g.pixels, PIXEL{4, -3, 1})
			g.pixels = append(g.pixels, PIXEL{4, -4, 1})
			g.pixels = append(g.pixels, PIXEL{5, -2, 1})

			g.pixels = append(g.pixels, PIXEL{9, -2, 1})
			g.pixels = append(g.pixels, PIXEL{10, -2, 1})
			g.pixels = append(g.pixels, PIXEL{10, -3, 1})
			g.pixels = append(g.pixels, PIXEL{10, -4, 1})

			g.pixels = append(g.pixels, PIXEL{4, 8, 1})
			g.pixels = append(g.pixels, PIXEL{4, 9, 1})
			g.pixels = append(g.pixels, PIXEL{4, 10, 1})
			g.pixels = append(g.pixels, PIXEL{5, 8, 1})

			g.pixels = append(g.pixels, PIXEL{9, 8, 1})
			g.pixels = append(g.pixels, PIXEL{10, 9, 1})
			g.pixels = append(g.pixels, PIXEL{10, 10, 1})
			g.pixels = append(g.pixels, PIXEL{10, 8, 1})

			g.pixels = append(g.pixels, PIXEL{12, 0, 1})
			g.pixels = append(g.pixels, PIXEL{13, 0, 1})
			g.pixels = append(g.pixels, PIXEL{14, 0, 1})
			g.pixels = append(g.pixels, PIXEL{12, 1, 1})

			g.pixels = append(g.pixels, PIXEL{12, 6, 1})
			g.pixels = append(g.pixels, PIXEL{13, 6, 1})
			g.pixels = append(g.pixels, PIXEL{14, 6, 1})
			g.pixels = append(g.pixels, PIXEL{12, 5, 1})

			g.pixels = append(g.pixels, PIXEL{4, 1, 1})
			g.pixels = append(g.pixels, PIXEL{4, 2, 1})
			g.pixels = append(g.pixels, PIXEL{5, 2, 1})
			g.pixels = append(g.pixels, PIXEL{5, 0, 1})
			g.pixels = append(g.pixels, PIXEL{6, 0, 1})
			g.pixels = append(g.pixels, PIXEL{6, 1, 1})

			g.pixels = append(g.pixels, PIXEL{8, 0, 1})
			g.pixels = append(g.pixels, PIXEL{8, 1, 1})
			g.pixels = append(g.pixels, PIXEL{9, 0, 1})
			g.pixels = append(g.pixels, PIXEL{10, 1, 1})
			g.pixels = append(g.pixels, PIXEL{10, 2, 1})
			g.pixels = append(g.pixels, PIXEL{9, 2, 1})

			g.pixels = append(g.pixels, PIXEL{4, 4, 1})
			g.pixels = append(g.pixels, PIXEL{4, 5, 1})
			g.pixels = append(g.pixels, PIXEL{5, 4, 1})
			g.pixels = append(g.pixels, PIXEL{6, 5, 1})
			g.pixels = append(g.pixels, PIXEL{6, 6, 1})
			g.pixels = append(g.pixels, PIXEL{5, 6, 1})

			g.pixels = append(g.pixels, PIXEL{8, 5, 1})
			g.pixels = append(g.pixels, PIXEL{8, 6, 1})
			g.pixels = append(g.pixels, PIXEL{9, 6, 1})
			g.pixels = append(g.pixels, PIXEL{9, 4, 1})
			g.pixels = append(g.pixels, PIXEL{10, 4, 1})
			g.pixels = append(g.pixels, PIXEL{10, 5, 1})
		}),
	)

	// BLINKER
	blinker_image, _ := loadButtonImage("patterns/blinker.png")
	button_blinker := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			}),
		),

		// specify the images to use
		widget.ButtonOpts.Image(blinker_image),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// BLINKER
			clear(g.pixels)
			g.is_figure_draw = true
			g.pixels = append(g.pixels, PIXEL{0, 0, 1})
			g.pixels = append(g.pixels, PIXEL{0, 1, 1})
			g.pixels = append(g.pixels, PIXEL{0, -1, 1})
		}),
	)

	// TOAD
	toad_image, _ := loadButtonImage("patterns/toad.png")
	button_toad := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			}),
		),

		// specify the images to use
		widget.ButtonOpts.Image(toad_image),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// TOAD
			clear(g.pixels)
			g.is_figure_draw = true
			g.pixels = append(g.pixels, PIXEL{0, 0, 1})
			g.pixels = append(g.pixels, PIXEL{0, 1, 1})
			g.pixels = append(g.pixels, PIXEL{1, 2, 1})
			g.pixels = append(g.pixels, PIXEL{2, -1, 1})
			g.pixels = append(g.pixels, PIXEL{3, 0, 1})
			g.pixels = append(g.pixels, PIXEL{3, 1, 1})
		}),
	)

	// add the button as a child of the container
	rootContainer.AddChild(button_block)
	rootContainer.AddChild(button_glider)
	rootContainer.AddChild(button_pulsar)
	rootContainer.AddChild(button_blinker)
	rootContainer.AddChild(button_toad)

	// construct the UI
	_ui := ebitenui.UI{
		Container: rootContainer,
	}
	g.ui = &_ui
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
	// обрабатываем нажатия
	g.keyEvent()

	// update the UI
	g.ui.Update()

	// рисуем пиксели, если нарисовали в игровой зоне
	mx, my := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.is_figure_draw {
			g.paintFigure(g.pixels, mx, my)

		} else {
			g.paint(mx, my)
		}
	}
	g.cursor = POS{
		x: mx,
		y: my,
	}

	return nil
}

func (g *MyGame) keyEvent() {
	// полная пауза
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.is_pause = true
		g.counter = -1
	}

	// если пауза, то не обновляем игру
	if !g.is_pause {
		g.counter++
	}

	// скорость 20
	if ebiten.IsKeyPressed(ebiten.Key1) {
		g.is_pause = false
		g.max_counter = 20
	}

	// скорость 10
	if ebiten.IsKeyPressed(ebiten.Key2) {
		g.is_pause = false
		g.max_counter = 10
	}

	// скорость 0
	if ebiten.IsKeyPressed(ebiten.Key3) {
		g.is_pause = false
		g.max_counter = 0
	}

	// переходим к следующему поколению
	if g.counter >= g.max_counter {
		g.field = gen_new_generation(g.height, g.width, g.field)
		g.counter = 0
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.y_offset++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.y_offset--

		// тут увеличиваем размер массива с помощью функции prepend
		// и зануляем отклонение, чтобы не сломать массив
		if g.y_offset < 0 {
			g.field = append(g.field)
			empty_arr := make([]byte, 1, 1)
			for i := 0; i < g.height; i++ {
				g.field[i] = append(empty_arr, g.field[i]...)
			}

			// fmt.Println("width = ", len(g.field[0]), "   g.y_offset = ", g.y_offset)
			g.width++
			g.y_offset = 0
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.x_offset--

		// тут увеличиваем размер массива с помощью функции prepend
		// и зануляем отклонение, чтобы не сломать массив
		if g.x_offset < 0 {
			empty_arr := make([][]byte, 0, 0)
			sub_arr := make([]byte, g.width, g.width)
			empty_arr = append(empty_arr, sub_arr)
			g.field = append(empty_arr, g.field...)

			// fmt.Println("height = ", len(g.field), "   g.x_offset = ", g.x_offset)
			g.height++
			g.x_offset = 0
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.x_offset++
	}
}

// paint draws the brush on the given canvas image at the position (x, y).
func (g *MyGame) paint(x, y int) {
	if x < gameHeight*4 && y < gameWidth*4 && x > 0 && y > 0 {
		g.field[g.x_offset+(x/4)][g.y_offset+(y/4)] = 1
	}
}

type PIXEL struct {
	x     int
	y     int
	value byte
}

func (g *MyGame) paintFigure(pixels []PIXEL, x, y int) {
	var loc_x = g.x_offset + x/4
	var loc_y = g.y_offset + y/4

	for _, pix := range pixels {
		if pix.x+loc_x > g.x_offset+gameHeight-1 || pix.x+loc_x < 0 || pix.y+loc_y > g.y_offset+gameWidth-1 || pix.y+loc_y < 0 {
			continue
		}
		g.field[pix.x+loc_x][pix.y+loc_y] = pix.value

	}
	g.is_figure_draw = false
}

func loadButtonImage(filename string) (*widget.ButtonImage, error) {
	var img, _, _ = ebitenutil.NewImageFromFile(filename)
	idle := image.NewNineSliceSimple(img, 1, 50)
	hover := image.NewNineSliceSimple(img, 1, 50)
	pressed := image.NewNineSliceSimple(img, 1, 50)

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *MyGame) Draw(screen *ebiten.Image) {

	// очищаем экран
	screen.Fill(white)
	// screen.DrawImage(g.canvasImage, nil)

	// показываем подсказки об управлении
	showHints(screen)

	// screen.SubImage()
	// draw the UI onto the screen
	g.ui.Draw(screen)

	// рисуем линии отделяющие шаблонные фигуры
	ebitenutil.DrawRect(screen, screenWidth-60, 0, 10, screenHeight, color.Black)

	var x_size = gameHeight + g.x_offset
	var y_size = gameWidth + g.y_offset

	// если вышли за границу массива, стоит его расширить и забить нулями
	if x_size > g.height {
		for i := 0; i < x_size-g.height; i++ {
			g.field = append(g.field, make([]byte, g.width, g.width))
		}
		g.height = x_size
		// fmt.Println("new height = ", g.height)
	}

	if y_size > g.width {
		var delta = y_size - g.width
		for i := 0; i < g.height; i++ {
			for range delta {
				g.field[i] = append(g.field[i], 0)
			}
		}
		g.width = y_size
		// fmt.Println("new width = ", g.width)
	}

	for x := g.x_offset; x < x_size; x++ {
		// fmt.Println("height = ", len(g.field))
		// fmt.Println("width = ", len(g.field[0]))
		// fmt.Println("x_size = ", x_size) // 1640
		for y := g.y_offset; y < y_size; y++ {
			if g.field[x][y] == 1 {
				for x1 := 0; x1 < scale; x1++ {
					for y1 := 0; y1 < scale; y1++ {
						screen.Set(((((x - g.x_offset) % gameHeight) * scale) + x1),
							((((y - g.y_offset) % gameWidth) * scale) + y1), black)
					}
				}
			}
		}
	}
}

func showHints(screen *ebiten.Image) {
	// Draw the message.
	tutorial := "Space: Pause\nArrow to move\n1, 2, 3: New generation frequency (1 - slow, 3 - fast)"
	msg := fmt.Sprintf("%s", tutorial)
	ebitenutil.DebugPrint(screen, msg)
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
	if err := ebiten.RunGame(NewGame(0)); err != nil {
		log.Fatal(err)
	}
}
