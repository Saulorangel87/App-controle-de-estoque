package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Conectar() {
	var err error
	DB, err = sql.Open("sqlite", "./estoque.db")
	if err != nil {
		log.Fatal("erro ao abrir banco:", err)
	}

	criarTabelas()
}

func criarTabelas() {
	usuarios := `
	CREATE TABLE IF NOT EXISTS usuarios (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT UNIQUE NOT NULL,
		senha_hash TEXT NOT NULL
	);`

	itens := `
	CREATE TABLE IF NOT EXISTS itens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		usuario_id INTEGER NOT NULL,
		nome TEXT NOT NULL,
		quantidade REAL NOT NULL,
		unidade TEXT NOT NULL,
		local TEXT NOT NULL,
		estoque_minimo REAL NOT NULL DEFAULT 0,
		FOREIGN KEY (usuario_id) REFERENCES usuarios(id)
	);`

	if _, err := DB.Exec(usuarios); err != nil {
		log.Fatal("erro ao criar tabela usuarios:", err)
	}
	if _, err := DB.Exec(itens); err != nil {
		log.Fatal("erro ao criar tabela itens:", err)
	}
}
