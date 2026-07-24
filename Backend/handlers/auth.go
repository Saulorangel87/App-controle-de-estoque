package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"controle-estoque/config"
	"controle-estoque/database"
	"controle-estoque/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type credenciais struct {
	Nome  string `json:"nome"`
	Senha string `json:"senha"`
}

func Cadastrar(w http.ResponseWriter, r *http.Request) {
	var c credenciais
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if c.Nome == "" || c.Senha == "" {
		http.Error(w, "nome e senha são obrigatórios", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(c.Senha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec(
		"INSERT INTO usuarios (nome, senha_hash) VALUES (?, ?)",
		c.Nome, string(hash),
	)
	if err != nil {
		http.Error(w, "usuário já existe", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "usuário criado"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var c credenciais
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	var usuario models.Usuario
	row := database.DB.QueryRow(
		"SELECT id, nome, senha_hash FROM usuarios WHERE nome = ?", c.Nome,
	)
	if err := row.Scan(&usuario.ID, &usuario.Nome, &usuario.SenhaHash); err != nil {
		http.Error(w, "usuário ou senha inválidos", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(c.Senha)); err != nil {
		http.Error(w, "usuário ou senha inválidos", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usuario_id": usuario.ID,
		"nome":       usuario.Nome,
		"exp":        time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenAssinado, err := token.SignedString(config.ChaveSecreta)
	if err != nil {
		http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenAssinado})
}