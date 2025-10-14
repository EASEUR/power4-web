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

// Game contient l'état du jeu
type Game struct {
	Grid    [ROWS][COLS]int // 0 = vide, 1 = joueur 1 (rouge), 2 = joueur 2 (jaune)
	Player  int             // joueur courant (1 ou 2)
	Winner  int             // 0 = pas de gagnant, 1 ou 2 = gagnant
	LastRow int
	LastCol int
}

var (
	game Game
	mu   sync.Mutex
	tmpl = template.Must(template.ParseFiles("templates/index.html"))
)

// playMove insère une pièce dans la colonne col (0..6).
// Retourne true si l'insertion a réussi, false si colonne invalide/pleine.
func (g *Game) playMove(col int) bool {
	if col < 0 || col >= COLS {
		return false
	}
	for r := ROWS - 1; r >= 0; r-- { // du bas vers le haut
		if g.Grid[r][col] == 0 {
			g.Grid[r][col] = g.Player
			g.LastRow = r
			g.LastCol = col
			return true
		}
	}
	return false
}

// switchPlayer alterne le joueur courant
func (g *Game) switchPlayer() {
	if g.Player == 1 {
		g.Player = 2
	} else {
		g.Player = 1
	}
}

// checkWin parcourt la grille et retourne le gagnant (1 ou 2) ou 0 si aucun.
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
	// diagonale descendante (\)
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
	// diagonale montante (/)
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

// reset remet la partie à zéro
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

// homeHandler affiche la page principale.
// On prépare une grille "affichage" ([][]string) où chaque case vaut:
// "" (vide), "red" ou "yellow" — le template utilise ces classes CSS.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	// construire une grille de strings pour le template
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

	// exécuter le template
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// playHandler traite les POST /play avec field "column".
// On accepte des formulaires qui envoient "column" en POST.
func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// parse form values (utile si body form-encoded)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil {
		// valeur invalide -> on redirige, sans rien faire
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	mu.Lock()
	// si partie déjà finie, on ignore les coups (ou tu peux reset si tu préfères)
	if game.Winner == 0 {
		if game.playMove(col) {
			// vérifier victoire
			if winner := game.checkWin(); winner != 0 {
				game.Winner = winner
			} else {
				game.switchPlayer()
			}
		}
	}
	mu.Unlock()

	// redirige toujours vers la page principale pour voir l'état mis à jour
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// resetHandler permet de réinitialiser la partie (POST /reset)
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

func main() {
	// initialisation du jeu
	game.reset()

	// handlers
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/reset", resetHandler)

	// servir CSS / static files depuis ./static/
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := ":83" // tu as demandé le port 83
	fmt.Println("Serveur lancé sur http://localhost" + port)
	// Attention : si tu es sur Unix, l'écoute sur un port <1024 demande souvent des droits root.
	// Si tu veux éviter d'utiliser sudo, change ici en ":8080" et va sur http://localhost:8080
	log.Fatal(http.ListenAndServe(port, nil))
}
