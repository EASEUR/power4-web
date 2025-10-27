package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

const ROWS = 6
const COLS = 7

// Cette structure contient les variables principales qui sont nécessaires à la jouabilité
// (qu'on enverra ensuite au template html via la fonction homeHandler)
type Game struct {
	Grid    [ROWS][COLS]int // état des cellules (0 = vide, 1= joueur 1, 2= joueur 2)
	Player  int             // joueur actuel qui doit jouer
	Winner  int             // détermine le gagnant
	LastRow int             //
	LastCol int             // Ces 2 variables (LastCol et LastRow) ont été rajoutées pour le style d'effet de "drop" des jetons
	// (pas nécessaire pour le jeu mais ajoute du style)
}

var (
	game Game
	mu   sync.Mutex
	tmpl = template.Must(template.ParseFiles("templates/index.html"))
)

// playMove pour jouer un coup, vérifie si la colonne est pleine
func (g *Game) playMove(col int) bool {
	if col < 0 || col >= COLS {
		return false
	}
	for r := ROWS - 1; r >= 0; r-- { // boucle pour faire descendre le jeton jusqu'en bas
		if g.Grid[r][col] == 0 {
			g.Grid[r][col] = g.Player
			g.LastRow = r
			g.LastCol = col
			return true
		}
	}
	return false
}

// switchPlayer pour alterner le tour des joueurs
func (g *Game) switchPlayer() {
	if g.Player == 1 {
		g.Player = 2
	} else {
		g.Player = 1
	}
}

// checkWin vérifie s'il y a un gagnant avec 4 jetons alignés
func (g *Game) checkWin() int {
	// horizontal
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS-3; c++ {
			v := g.Grid[r][c]
			if v != 0 &&
				g.Grid[r][c+1] == v &&
				g.Grid[r][c+2] == v &&
				g.Grid[r][c+3] == v {
				return v
			}
		}
	}
	// vertical
	for r := 0; r < ROWS-3; r++ {
		for c := 0; c < COLS; c++ {
			v := g.Grid[r][c]
			if v != 0 &&
				g.Grid[r+1][c] == v &&
				g.Grid[r+2][c] == v &&
				g.Grid[r+3][c] == v {
				return v
			}
		}
	}
	// diagonale descendante
	for r := 0; r < ROWS-3; r++ {
		for c := 0; c < COLS-3; c++ {
			v := g.Grid[r][c]
			if v != 0 &&
				g.Grid[r+1][c+1] == v &&
				g.Grid[r+2][c+2] == v &&
				g.Grid[r+3][c+3] == v {
				return v
			}
		}
	}
	// diagonale montante
	for r := 3; r < ROWS; r++ {
		for c := 0; c < COLS-3; c++ {
			v := g.Grid[r][c]
			if v != 0 &&
				g.Grid[r-1][c+1] == v &&
				g.Grid[r-2][c+2] == v &&
				g.Grid[r-3][c+3] == v {
				return v
			}
		}
	}
	return 0
}

// option reset pour remettre la partie à 0
func (g *Game) reset() {
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			g.Grid[r][c] = 0
		}
	}
	g.Player = 1
	g.Winner = 0
	g.LastRow = -1
	g.LastCol = -1

}

// HomeHandler pour afficher la page principale, création du tableau de puissance 4 pour l'envoyer au template html
func homeHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	display := make([][]string, ROWS)
	for i := 0; i < ROWS; i++ {
		display[i] = make([]string, COLS)
		for j := 0; j < COLS; j++ {
			switch game.Grid[i][j] {
			case 0:
				display[i][j] = "empty"
			case 1:
				display[i][j] = "red"
			case 2:
				display[i][j] = "yellow"
			}
		}
	}
	// les variables de data sont les variables que l'on veut récupérer dans le fichier html
	// Donc ici on a les variables de cellules, tour du joueur, gagnant, et les 2 variables pour l'effet
	//de drop des jetons
	data := struct {
		Grid    [][]string
		Player  int
		Winner  int
		LastRow int
		LastCol int
	}{
		Grid:    display,
		Player:  game.Player,
		Winner:  game.Winner,
		LastRow: game.LastRow,
		LastCol: game.LastCol,
	}
	mu.Unlock()

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// playHandler pour update le coup joué vers le site
func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// vérifie la win
	mu.Lock()
	if game.Winner == 0 {
		if game.playMove(col) {
			if winner := game.checkWin(); winner != 0 {
				game.Winner = winner
			} else {
				game.switchPlayer()
			}
		}
	}
	mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther) // redirige toujours vers la page principale
}

// resetHandler pour reset la partie en cours
func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	mu.Lock()
	game.reset()
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// La fonction main reprend toutes les fonctions de handler nécessaires à la jouabilité pour les envoyer sur le site
// on détermine aussi le chemin d'accès au fichier CSS
// et le port de réseau utilisé pour héberger localement
func main() {

	game.reset()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/reset", resetHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := ":3333"
	fmt.Println("Serveur lancé sur http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
