package database

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Conectar() {
	// Em produção, DB_PATH aponta para dentro do volume montado (ex: /app/data/estoque.db),
	// assim o arquivo sobrevive a rebuilds do container. Em desenvolvimento local, sem essa
	// variável definida, cai de volta pro caminho relativo de sempre.
	caminho := os.Getenv("DB_PATH")
	if caminho == "" {
		caminho = "./estoque.db"
	}

	var err error
	DB, err = sql.Open("sqlite", caminho)
	if err != nil {
		log.Fatal("erro ao abrir banco:", err)
	}

	criarTabelas()
	migrarColunasSeguranca()
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

// migrarColunasSeguranca adiciona as colunas de pergunta/resposta de segurança
// em bancos que já existiam antes dessa funcionalidade (CREATE TABLE IF NOT EXISTS
// não adiciona colunas novas numa tabela que já existe, então isso é feito à parte).
// Verifica antes se a coluna já existe, para não dar erro rodando de novo toda vez.
func migrarColunasSeguranca() {
	if !colunaExiste("usuarios", "pergunta_seguranca") {
		if _, err := DB.Exec(`ALTER TABLE usuarios ADD COLUMN pergunta_seguranca TEXT NOT NULL DEFAULT ''`); err != nil {
			log.Fatal("erro ao migrar coluna pergunta_seguranca:", err)
		}
	}
	if !colunaExiste("usuarios", "resposta_seguranca_hash") {
		if _, err := DB.Exec(`ALTER TABLE usuarios ADD COLUMN resposta_seguranca_hash TEXT NOT NULL DEFAULT ''`); err != nil {
			log.Fatal("erro ao migrar coluna resposta_seguranca_hash:", err)
		}
	}
}

// colunaExiste consulta o esquema da tabela (PRAGMA table_info) para saber se
// uma coluna específica já existe, evitando tentar adicioná-la duas vezes.
func colunaExiste(tabela, coluna string) bool {
	rows, err := DB.Query("PRAGMA table_info(" + tabela + ")")
	if err != nil {
		log.Fatal("erro ao verificar colunas:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var nome, tipo string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &nome, &tipo, &notNull, &dflt, &pk); err != nil {
			log.Fatal("erro ao ler colunas:", err)
		}
		if nome == coluna {
			return true
		}
	}
	return false
}
