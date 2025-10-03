package main

import (
	"fmt"
)

const ROWS = 6
const COLS = 7

func createBoard() [][]string {
	board := make([][]string, ROWS)
	for i := range board {
		board[i] = make([]string, COLS)
		for j := range board[i] {
			board[i][j] = "."
		}
	}
	return board
}

func printBoard(board [][]string) {
	for _, row := range board {
		for _, cell := range row {
			fmt.Print(cell, " ")
		}
		fmt.Println()
	}
	for i := 0; i < COLS; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()
}

func dropPiece(board [][]string, col int, piece string) bool {
	for i := ROWS - 1; i >= 0; i-- {
		if board[i][col] == "." {
			board[i][col] = piece
			return true
		}
	}
	return false // colonne pleine
}

func checkVictory(board [][]string, piece string) bool {
	// Vérifie horizontal
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS-3; c++ {
			if board[r][c] == piece &&
				board[r][c+1] == piece &&
				board[r][c+2] == piece &&
				board[r][c+3] == piece {
				return true
			}
		}
	}
	// Vérifie vertical
	for r := 0; r < ROWS-3; r++ {
		for c := 0; c < COLS; c++ {
			if board[r][c] == piece &&
				board[r+1][c] == piece &&
				board[r+2][c] == piece &&
				board[r+3][c] == piece {
				return true
			}
		}
	}
	// Vérifie diagonale descendante
	for r := 0; r < ROWS-3; r++ {
		for c := 0; c < COLS-3; c++ {
			if board[r][c] == piece &&
				board[r+1][c+1] == piece &&
				board[r+2][c+2] == piece &&
				board[r+3][c+3] == piece {
				return true
			}
		}
	}
	// Vérifie diagonale montante
	for r := 3; r < ROWS; r++ {
		for c := 0; c < COLS-3; c++ {
			if board[r][c] == piece &&
				board[r-1][c+1] == piece &&
				board[r-2][c+2] == piece &&
				board[r-3][c+3] == piece {
				return true
			}
		}
	}
	return false
}

func isFull(board [][]string) bool {
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			if board[r][c] == "." {
				return false
			}
		}
	}
	return true
}

func main() {
	board := createBoard()
	turn := 0
	gameOver := false

	for !gameOver {
		printBoard(board)
		var col int
		fmt.Printf("Joueur %d, choisis une colonne (0-6) : ", (turn%2)+1)
		fmt.Scan(&col)

		piece := "X"
		if turn%2 == 1 {
			piece = "O"
		}

		if col >= 0 && col < COLS && dropPiece(board, col, piece) {
			if checkVictory(board, piece) {
				printBoard(board)
				fmt.Printf("Joueur %d a gagné !\n", (turn%2)+1)
				gameOver = true
			} else if isFull(board) {
				printBoard(board)
				fmt.Println("Match nul !")
				gameOver = true
			} else {
				turn++
			}
		} else {
			fmt.Println("Colonne invalide, réessaye.")
		}
	}
}
