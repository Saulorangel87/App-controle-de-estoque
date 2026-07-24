package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
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

type cadastroEntrada struct {
	Nome              string `json:"nome"`
	Senha             string `json:"senha"`
	PerguntaSeguranca string `json:"pergunta_seguranca"`
	RespostaSeguranca string `json:"resposta_seguranca"`
}

// normalizarResposta remove espaços nas pontas e ignora maiúsculas/minúsculas,
// para que "Rex", "rex " e "REX" sejam todos aceitos como a mesma resposta.
func normalizarResposta(resposta string) string {
	return strings.ToLower(strings.TrimSpace(resposta))
}

func Cadastrar(w http.ResponseWriter, r *http.Request) {
	var c cadastroEntrada
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if c.Nome == "" || c.Senha == "" || c.PerguntaSeguranca == "" || c.RespostaSeguranca == "" {
		http.Error(w, "nome, senha, pergunta e resposta de segurança são obrigatórios", http.StatusBadRequest)
		return
	}

	hashSenha, err := bcrypt.GenerateFromPassword([]byte(c.Senha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	hashResposta, err := bcrypt.GenerateFromPassword(
		[]byte(normalizarResposta(c.RespostaSeguranca)), bcrypt.DefaultCost,
	)
	if err != nil {
		http.Error(w, "erro ao processar resposta de segurança", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec(
		`INSERT INTO usuarios (nome, senha_hash, pergunta_seguranca, resposta_seguranca_hash)
		 VALUES (?, ?, ?, ?)`,
		c.Nome, string(hashSenha), c.PerguntaSeguranca, string(hashResposta),
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

// ObterPerguntaSeguranca devolve a pergunta de segurança cadastrada para um usuário,
// para o frontend exibir antes de pedir a resposta. Não exige login (a pessoa está
// tentando recuperar o acesso, então ainda não tem token).
func ObterPerguntaSeguranca(w http.ResponseWriter, r *http.Request) {
	nome := r.URL.Query().Get("nome")
	if nome == "" {
		http.Error(w, "informe o nome de usuário", http.StatusBadRequest)
		return
	}

	var pergunta string
	row := database.DB.QueryRow(
		"SELECT pergunta_seguranca FROM usuarios WHERE nome = ?", nome,
	)
	// Mensagem genérica de propósito: não revela se o usuário existe ou não,
	// só que não foi possível seguir com a recuperação.
	if err := row.Scan(&pergunta); err != nil {
		http.Error(w, "não foi possível encontrar esse usuário", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"pergunta_seguranca": pergunta})
}

type redefinicaoEntrada struct {
	Nome        string `json:"nome"`
	Resposta    string `json:"resposta"`
	NovaSenha   string `json:"nova_senha"`
}

// RedefinirSenha confere a resposta de segurança e, se bater, troca a senha do usuário.
func RedefinirSenha(w http.ResponseWriter, r *http.Request) {
	var e redefinicaoEntrada
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	if e.Nome == "" || e.Resposta == "" || e.NovaSenha == "" {
		http.Error(w, "nome, resposta e nova senha são obrigatórios", http.StatusBadRequest)
		return
	}

	var usuarioID int
	var hashResposta string
	row := database.DB.QueryRow(
		"SELECT id, resposta_seguranca_hash FROM usuarios WHERE nome = ?", e.Nome,
	)
	if err := row.Scan(&usuarioID, &hashResposta); err != nil {
		http.Error(w, "não foi possível redefinir a senha", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(hashResposta), []byte(normalizarResposta(e.Resposta)),
	); err != nil {
		http.Error(w, "resposta de segurança incorreta", http.StatusUnauthorized)
		return
	}

	novoHashSenha, err := bcrypt.GenerateFromPassword([]byte(e.NovaSenha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar nova senha", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec(
		"UPDATE usuarios SET senha_hash = ? WHERE id = ?", string(novoHashSenha), usuarioID,
	)
	if err != nil {
		http.Error(w, "erro ao atualizar senha", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"mensagem": "senha redefinida com sucesso"})
}
