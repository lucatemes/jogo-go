package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// Define os elementos do jogo
type Elemento struct {
	simbolo  rune
	cor      termbox.Attribute
	corFundo termbox.Attribute
	tangivel bool
}

// jogador controlado pelo jogador
var jogador = Elemento{
	simbolo:  '☺',
	cor:      termbox.ColorBlack,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var moeda = Elemento{
	simbolo:  '🪙',
	cor:      termbox.ColorYellow,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

// Parede
var parede = Elemento{
	simbolo:  '▤',
	cor:      termbox.ColorLightMagenta | termbox.AttrBold | termbox.AttrDim,
	corFundo: termbox.ColorDarkGray,
	tangivel: true,
}

// Barrreira
var barreira = Elemento{
	simbolo:  '#',
	cor:      termbox.ColorRed,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

// Vegetação
var vegetacao = Elemento{
	simbolo:  '♣',
	cor:      termbox.ColorGreen,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

// inimigo
var inimigo = Elemento{
	simbolo:  '☠',
	cor:      termbox.ColorRed,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

// Elemento vazio
var vazio = Elemento{
	simbolo:  ' ',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var especial = Elemento{
	simbolo:  '*',
	cor:      termbox.ColorLightBlue,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

// Elemento para representar áreas não reveladas (efeito de neblina)
var neblina = Elemento{
	simbolo:  '.',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorYellow,
	tangivel: false,
}

var mapa [][]Elemento
var posX, posY int
var ultimoElementoSobjogador = vazio
var statusMsg string
var ultimoElementoSobPersonagem Elemento

var posX1, posY1 int
var pontos int
var pontosMax int

var inimigoCongelado bool

var efeitoNeblina = false
var revelado [][]bool
var raioVisao int = 3
var mapaMutex = &sync.Mutex{}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	carregarMapa("mapa.txt")
	if efeitoNeblina {
		revelarArea()
	}
	desenhaTudo()
	pontosMax = 15

	go func() {
		for {
			perseguirJogador(inimigo, jogador, parede)
			moverMoeda(moeda)
		}
	}()
	go desenhaTudo()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return // Sair do programa
			}
			if ev.Ch == 'f' {
				interagir(jogador, moeda)
				congelarInimigo(especial, inimigo)
			} else {
				mover(ev.Ch)
				if efeitoNeblina {
					revelarArea()
				}
			}
			desenhaTudo()
		}
	}
}

func calcularPontosMax() int {
	pontosMax := 0
	for y := 0; y < len(mapa); y++ {
		for x := 0; x < len(mapa[y]); x++ {
			if reflect.DeepEqual(mapa[y][x], moeda) {
				pontosMax++
			}
		}
	}
	return pontosMax
}

func checarMoedas(pontos int) {
	if pontos == pontosMax {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		termbox.Flush()
		time.Sleep(time.Second)
		os.Exit(0)
		return
	}
}

func moverMoeda(moeda Elemento) {
	mapaMutex.Lock()
	defer mapaMutex.Unlock()
	for y := 0; y < len(mapa); y++ {
		for x := 0; x < len(mapa[y]); x++ {
			if reflect.DeepEqual(mapa[y][x], moeda) {
				mapa[y][x] = vazio
				rand.Seed(time.Now().UnixNano())
				num := rand.Intn(3) - 1 // Gera -1, 0 ou 1
				mapa[y+num][x+num] = moeda
			}
		}
		time.Sleep(100 * time.Millisecond)
		desenhaTudo()
	}

}

func perseguirJogador(inimigo, jogador, parede Elemento) {
	for {
		mapaMutex.Lock()
		defer mapaMutex.Unlock()
		posXInimigo, posYInimigo := buscarPosicaoInimigo(inimigo)
		posXJogador, posYJogador := buscarPosicaoInimigo(jogador)

		// verifica se o inimigo já está na mesma posição que o jogador
		if posXInimigo == posXJogador && posYInimigo == posYJogador {
			telaDerrota(pontos)
			return
		}

		// Determina a nova posição do inimigo
		novaPosX, novaPosY := posXInimigo, posYInimigo
		if posXInimigo < posXJogador && posXInimigo+1 < len(mapa[0]) && mapa[posYInimigo][posXInimigo+1] != parede {
			novaPosX++
		} else if posXInimigo > posXJogador && posXInimigo-1 >= 0 && mapa[posYInimigo][posXInimigo-1] != parede {
			novaPosX--
		}
		if posYInimigo < posYJogador && posYInimigo+1 < len(mapa) && mapa[posYInimigo+1][posXInimigo] != parede {
			novaPosY++
		} else if posYInimigo > posYJogador && posYInimigo-1 >= 0 && mapa[posYInimigo-1][posXInimigo] != parede {
			novaPosY--
		}

		// Check if the new positions are within the bounds of the mapa array
		if novaPosX >= 0 && novaPosX < len(mapa[0]) && novaPosY >= 0 && novaPosY < len(mapa) {
			// Mova o inimigo para a nova posição
			mapa[posYInimigo][posXInimigo] = vazio
			mapa[novaPosY][novaPosX] = inimigo
		}

		mapaMutex.Unlock()
		desenhaTudo()
		time.Sleep(200 * time.Millisecond)
	}
}

func telaDerrota(pontos int) {
	fmt.Println("Game Over")
	fmt.Printf("Pontos: %d\n", pontos)
	os.Exit(0)
}

func carregarMapa(nomeArquivo string) {
	arquivo, err := os.Open(nomeArquivo)
	if err != nil {
		panic(err)
	}
	defer arquivo.Close()

	scanner := bufio.NewScanner(arquivo)
	y := 0

	for scanner.Scan() {
		linhaTexto := scanner.Text()
		var linhaElementos []Elemento
		var linhaRevelada []bool
		for x, char := range linhaTexto {
			elementoAtual := vazio
			switch char {
			case parede.simbolo:
				elementoAtual = parede
			case barreira.simbolo:
				elementoAtual = barreira
			case vegetacao.simbolo:
				elementoAtual = vegetacao
			case inimigo.simbolo:
				elementoAtual = inimigo
			case moeda.simbolo:
				elementoAtual = moeda
			case especial.simbolo:
				elementoAtual = especial
			case jogador.simbolo:
				// Atualiza a posição inicial do jogador
				posX, posY = x, y
				elementoAtual = vazio
			}
			linhaElementos = append(linhaElementos, elementoAtual)
			linhaRevelada = append(linhaRevelada, false)
		}
		mapa = append(mapa, linhaElementos)
		revelado = append(revelado, linhaRevelada)
		y++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, linha := range mapa {
		for x, elem := range linha {
			if efeitoNeblina == false || revelado[y][x] {
				termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
			} else {
				termbox.SetCell(x, y, neblina.simbolo, neblina.cor, neblina.corFundo)
			}
		}
	}

	desenhaBarraDeStatus()

	termbox.Flush()
}

func desenhaBarraDeStatus() {
	for i, c := range statusMsg {
		termbox.SetCell(i, len(mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e F para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	}
}

func revelarArea() {
	minX := max(0, posX-raioVisao)
	maxX := min(len(mapa[0])-1, posX+raioVisao)
	minY := max(0, posY-raioVisao/2)
	maxY := min(len(mapa)-1, posY+raioVisao/2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Revela as células dentro do quadrado de visão
			revelado[y][x] = true
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buscarPosicaoInimigo(inimigo Elemento) (int, int) {
	for y := 0; y < len(mapa); y++ {
		for x := 0; x < len(mapa[y]); x++ {
			if reflect.DeepEqual(mapa[y][x], inimigo) {
				return x, y
			}
		}
	}

	return -1, -1
}

func exibirPontos() {
	// seleciona local de exibicao
	fmt.Print("\033[0;0H")
	// limpa a linha atual
	fmt.Print("\033[K")
	// imprime a pontuação
	fmt.Printf("Pontos: %d", pontos)
}

func mover(comando rune) {
	novaPosX, novaPosY := posX, posY

	switch comando {
	case 'w':
		novaPosY--
	case 'a':
		novaPosX--
	case 's':
		novaPosY++
	case 'd':
		novaPosX++
	}

	// checa se a nova posição é válida
	if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[0]) && mapa[novaPosY][novaPosX] != parede && mapa[novaPosY][novaPosX] != moeda && mapa[novaPosY][novaPosX] != especial {
		mapa[posY][posX] = vazio
		posY, posX = novaPosY, novaPosX
		mapa[posY][posX] = jogador
	}
}

func interagir(jogador, moeda Elemento) {
	posX, posY := buscarPosicaoInimigo(jogador)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			x, y := posX+dx, posY+dy
			if y >= 0 && y < len(mapa) && x >= 0 && x < len(mapa[0]) {
				if reflect.DeepEqual(mapa[y][x], moeda) && dx*dx+dy*dy <= 8 {
					pontos++
					mapa[y][x] = vazio
					exibirPontos()
					checarMoedas(pontos)
				}

			}
		}
	}
}

// nao esta funcionando daqui pra baixo
func congelarInimigo(especial, inimigo Elemento) {
	posX, posY := buscarPosicaoInimigo(inimigo)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			x, y := posX+dx, posY+dy
			if y >= 0 && y < len(mapa) && x >= 0 && x < len(mapa[0]) {
				if reflect.DeepEqual(mapa[y][x], especial) {
					inimigoCongelado = true
					mudaCorInimigo(inimigo)
					desenhaTudo()
					go func() {
						mapa[posX][posY] = inimigo
						time.Sleep(3 * time.Second)
						inimigoCongelado = false
						desenhaTudo()
						mudaCorInimigo(inimigo)
					}()
				}
			}
		}
	}
}

func mudaCorInimigo(inimigo Elemento) {
	if inimigo.cor == termbox.ColorRed {
		inimigo.cor = termbox.ColorBlue
	} else {
		inimigo.cor = termbox.ColorRed
	}
}
