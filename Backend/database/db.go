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

	confirmacoesImportacao := `
	CREATE TABLE IF NOT EXISTS confirmacoes_importacao (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		usuario_id INTEGER NOT NULL,
		chave TEXT NOT NULL,
		atualizados INTEGER NOT NULL,
		criados INTEGER NOT NULL,
		criada_em TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (usuario_id, chave),
		FOREIGN KEY (usuario_id) REFERENCES usuarios(id)
	);`

	if _, err := DB.Exec(usuarios); err != nil {
		log.Fatal("erro ao criar tabela usuarios:", err)
	}
	if _, err := DB.Exec(itens); err != nil {
		log.Fatal("erro ao criar tabela itens:", err)
	}
	if _, err := DB.Exec(confirmacoesImportacao); err != nil {
		log.Fatal("erro ao criar tabela de confirmações:", err)
	}

	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS codigos_email (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			usuario_id INTEGER NOT NULL,
			finalidade TEXT NOT NULL,
			codigo_hash TEXT NOT NULL,
			expira_em TEXT NOT NULL,
			criado_em TEXT NOT NULL,
			tentativas INTEGER NOT NULL DEFAULT 0,
			usado_em TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (usuario_id) REFERENCES usuarios(id)
		)
	`); err != nil {
		log.Fatal("erro ao criar tabela de códigos de e-mail:", err)
	}
	if _, err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_codigos_email_usuario
		ON codigos_email (usuario_id, finalidade, id DESC)`); err != nil {
		log.Fatal("erro ao criar índice de códigos de e-mail:", err)
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
	if !colunaExiste("usuarios", "token_versao") {
		if _, err := DB.Exec(`ALTER TABLE usuarios ADD COLUMN token_versao INTEGER NOT NULL DEFAULT 0`); err != nil {
			log.Fatal("erro ao migrar coluna token_versao:", err)
		}
	}
	if !colunaExiste("usuarios", "email") {
		if _, err := DB.Exec(`ALTER TABLE usuarios ADD COLUMN email TEXT NOT NULL DEFAULT ''`); err != nil {
			log.Fatal("erro ao migrar coluna email:", err)
		}
	}
	if !colunaExiste("usuarios", "email_verificado_em") {
		if _, err := DB.Exec(`ALTER TABLE usuarios ADD COLUMN email_verificado_em TEXT NOT NULL DEFAULT ''`); err != nil {
			log.Fatal("erro ao migrar coluna email_verificado_em:", err)
		}
	}
	if _, err := DB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_usuarios_email
		ON usuarios (lower(email)) WHERE email <> ''`); err != nil {
		log.Fatal("erro ao criar índice de e-mail:", err)
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
