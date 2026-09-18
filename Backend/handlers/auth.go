package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"controle-estoque/config"
	"controle-estoque/database"
	"controle-estoque/middleware"
	"controle-estoque/models"
	"controle-estoque/services"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type credenciais struct {
	Nome  string `json:"nome"`
	Senha string `json:"senha"`
}

const nomeCookieSessao = "estoque_sessao"

type cadastroEntrada struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type codigoEmailEntrada struct {
	Nome   string `json:"nome"`
	Codigo string `json:"codigo"`
}

type emailEntrada struct {
	Email string `json:"email"`
}

type redefinicaoEntrada struct {
	Email     string `json:"email"`
	Codigo    string `json:"codigo"`
	NovaSenha string `json:"nova_senha"`
}

const (
	tamanhoMinimoSenha       = 8
	tamanhoMaximoSenhaBytes  = 72 // limite efetivo aceito pelo bcrypt
	tamanhoMaximoNomeUsuario = 100
	tamanhoMaximoEmail       = 254
	tamanhoCodigoEmail       = 8
	validadeCodigoEmail      = 15 * time.Minute
	intervaloNovoCodigo      = 60 * time.Second
	maxTentativasCodigo      = 5
	finalidadeVerificacao    = "verificacao"
	finalidadeRecuperacao    = "recuperacao"
)

var (
	errCodigoInvalido  = errors.New("código inválido ou expirado")
	errCodigoBloqueado = errors.New("código bloqueado")
	errCodigoEmEspera  = errors.New("um código recente ainda está válido")
)

func validarSenha(senha string) error {
	if utf8.RuneCountInString(senha) < tamanhoMinimoSenha {
		return errors.New("a senha deve ter pelo menos 8 caracteres")
	}
	if len([]byte(senha)) > tamanhoMaximoSenhaBytes {
		return errors.New("a senha excede o limite permitido")
	}
	return nil
}

func validarCampoAutenticacao(valor, campo string, tamanhoMaximo int) error {
	if strings.TrimSpace(valor) == "" {
		return errors.New(campo + " é obrigatório")
	}
	if utf8.RuneCountInString(valor) > tamanhoMaximo {
		return errors.New(campo + " excede o limite permitido")
	}
	return nil
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validarEmail(email string) error {
	email = normalizarEmail(email)
	if email == "" {
		return errors.New("e-mail é obrigatório")
	}
	if len([]byte(email)) > tamanhoMaximoEmail || strings.ContainsAny(email, "\r\n") {
		return errors.New("e-mail inválido")
	}
	endereco, err := mail.ParseAddress(email)
	if err != nil || endereco.Address != email || !strings.Contains(email, "@") {
		return errors.New("e-mail inválido")
	}
	return nil
}

func validarCodigo(codigo string) error {
	codigo = strings.TrimSpace(codigo)
	if len(codigo) != tamanhoCodigoEmail {
		return errors.New("código inválido")
	}
	for _, caractere := range codigo {
		if caractere < '0' || caractere > '9' {
			return errors.New("código inválido")
		}
	}
	return nil
}

func gerarCodigoEmail() (string, error) {
	numero, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", numero.Int64()), nil
}

func hashCodigo(codigo string) string {
	hash := sha256.Sum256([]byte(codigo))
	return hex.EncodeToString(hash[:])
}

func emitirCodigoEmail(usuarioID int, finalidade string) (string, error) {
	var criadoEm string
	err := database.DB.QueryRow(
		`SELECT criado_em FROM codigos_email
		 WHERE usuario_id = ? AND finalidade = ? AND usado_em = ''
		 ORDER BY id DESC LIMIT 1`,
		usuarioID, finalidade,
	).Scan(&criadoEm)
	if err == nil {
		if criado, parseErr := time.Parse(time.RFC3339Nano, criadoEm); parseErr == nil &&
			time.Since(criado) >= 0 && time.Since(criado) < intervaloNovoCodigo {
			return "", errCodigoEmEspera
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	codigo, err := gerarCodigoEmail()
	if err != nil {
		return "", err
	}
	agora := time.Now().UTC()
	tx, err := database.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE codigos_email SET usado_em = ?
		 WHERE usuario_id = ? AND finalidade = ? AND usado_em = ''`,
		agora.Format(time.RFC3339Nano), usuarioID, finalidade,
	); err != nil {
		return "", err
	}
	if _, err := tx.Exec(
		`INSERT INTO codigos_email
		 (usuario_id, finalidade, codigo_hash, expira_em, criado_em)
		 VALUES (?, ?, ?, ?, ?)`,
		usuarioID, finalidade, hashCodigo(codigo),
		agora.Add(validadeCodigoEmail).Format(time.RFC3339Nano),
		agora.Format(time.RFC3339Nano),
	); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return codigo, nil
}

func invalidarCodigoAtual(usuarioID int, finalidade string) {
	_, _ = database.DB.Exec(
		`UPDATE codigos_email SET usado_em = ?
		 WHERE usuario_id = ? AND finalidade = ? AND usado_em = ''`,
		time.Now().UTC().Format(time.RFC3339Nano), usuarioID, finalidade,
	)
}

func consumirCodigoEmail(usuarioID int, finalidade, codigo string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id, tentativas int
	var codigoHash, expiraEm, usadoEm string
	err = tx.QueryRow(
		`SELECT id, codigo_hash, expira_em, tentativas, usado_em
		 FROM codigos_email
		 WHERE usuario_id = ? AND finalidade = ?
		 ORDER BY id DESC LIMIT 1`,
		usuarioID, finalidade,
	).Scan(&id, &codigoHash, &expiraEm, &tentativas, &usadoEm)
	if err != nil || usadoEm != "" {
		return errCodigoInvalido
	}

	expiracao, err := time.Parse(time.RFC3339Nano, expiraEm)
	if err != nil || !time.Now().UTC().Before(expiracao) {
		return errCodigoInvalido
	}
	if tentativas >= maxTentativasCodigo {
		return errCodigoBloqueado
	}

	comparacao := subtle.ConstantTimeCompare([]byte(codigoHash), []byte(hashCodigo(codigo)))
	if comparacao != 1 {
		if _, err := tx.Exec("UPDATE codigos_email SET tentativas = tentativas + 1 WHERE id = ?", id); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return errCodigoInvalido
	}

	if _, err := tx.Exec(
		"UPDATE codigos_email SET usado_em = ? WHERE id = ?",
		time.Now().UTC().Format(time.RFC3339Nano), id,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func responderMensagem(w http.ResponseWriter, status int, mensagem string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"mensagem": mensagem})
}

func removerUsuarioComCodigos(usuarioID int) {
	tx, err := database.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM codigos_email WHERE usuario_id = ?", usuarioID); err != nil {
		return
	}
	if _, err := tx.Exec("DELETE FROM usuarios WHERE id = ?", usuarioID); err != nil {
		return
	}
	_ = tx.Commit()
}

func Cadastrar(w http.ResponseWriter, r *http.Request) {
	var c cadastroEntrada
	if err := decodificarJSON(r, &c); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}
	c.Nome = strings.TrimSpace(c.Nome)
	c.Email = normalizarEmail(c.Email)
	if err := validarCampoAutenticacao(c.Nome, "nome", tamanhoMaximoNomeUsuario); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validarEmail(c.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validarSenha(c.Senha); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !services.EmailResendConfigurado() {
		http.Error(w, "recuperação por e-mail ainda não está configurada", http.StatusServiceUnavailable)
		return
	}

	hashSenha, err := bcrypt.GenerateFromPassword([]byte(c.Senha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	resultado, err := database.DB.Exec(
		`INSERT INTO usuarios (nome, senha_hash, email, email_verificado_em)
		 VALUES (?, ?, ?, '')`,
		c.Nome, string(hashSenha), c.Email,
	)
	if err != nil {
		http.Error(w, "usuário ou e-mail já cadastrado", http.StatusConflict)
		return
	}
	usuarioID64, err := resultado.LastInsertId()
	if err != nil {
		http.Error(w, "erro ao criar usuário", http.StatusInternalServerError)
		return
	}
	usuarioID := int(usuarioID64)

	codigo, err := emitirCodigoEmail(usuarioID, finalidadeVerificacao)
	if err != nil {
		removerUsuarioComCodigos(usuarioID)
		http.Error(w, "não foi possível preparar a confirmação por e-mail", http.StatusInternalServerError)
		return
	}
	if err := services.EnviarCodigoResend(c.Email, codigo, finalidadeVerificacao); err != nil {
		removerUsuarioComCodigos(usuarioID)
		http.Error(w, "não foi possível enviar o e-mail de confirmação", http.StatusServiceUnavailable)
		return
	}

	responderMensagem(w, http.StatusCreated, "conta criada; informe o código enviado para seu e-mail")
}

func VerificarEmailCadastro(w http.ResponseWriter, r *http.Request) {
	var entrada codigoEmailEntrada
	if err := decodificarJSON(r, &entrada); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}
	entrada.Nome = strings.TrimSpace(entrada.Nome)
	entrada.Codigo = strings.TrimSpace(entrada.Codigo)
	if err := validarCampoAutenticacao(entrada.Nome, "nome", tamanhoMaximoNomeUsuario); err != nil || validarCodigo(entrada.Codigo) != nil {
		http.Error(w, "código inválido", http.StatusBadRequest)
		return
	}

	var usuarioID int
	var verificadoEm string
	err := database.DB.QueryRow(
		"SELECT id, email_verificado_em FROM usuarios WHERE nome = ?", entrada.Nome,
	).Scan(&usuarioID, &verificadoEm)
	if err != nil || verificadoEm != "" {
		http.Error(w, "código inválido ou expirado", http.StatusUnauthorized)
		return
	}
	if err := consumirCodigoEmail(usuarioID, finalidadeVerificacao, entrada.Codigo); err != nil {
		http.Error(w, "código inválido ou expirado", http.StatusUnauthorized)
		return
	}
	if _, err := database.DB.Exec(
		"UPDATE usuarios SET email_verificado_em = ? WHERE id = ?",
		time.Now().UTC().Format(time.RFC3339Nano), usuarioID,
	); err != nil {
		http.Error(w, "erro ao confirmar e-mail", http.StatusInternalServerError)
		return
	}
	responderMensagem(w, http.StatusOK, "e-mail confirmado")
}

func ReenviarVerificacaoEmail(w http.ResponseWriter, r *http.Request) {
	var entrada struct {
		Nome string `json:"nome"`
	}
	if err := decodificarJSON(r, &entrada); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}
	entrada.Nome = strings.TrimSpace(entrada.Nome)
	if err := validarCampoAutenticacao(entrada.Nome, "nome", tamanhoMaximoNomeUsuario); err != nil {
		responderMensagem(w, http.StatusOK, "se a conta estiver pendente, um novo código será enviado")
		return
	}
	var usuarioID int
	var email, verificadoEm string
	if err := database.DB.QueryRow(
		"SELECT id, email, email_verificado_em FROM usuarios WHERE nome = ?", entrada.Nome,
	).Scan(&usuarioID, &email, &verificadoEm); err != nil || verificadoEm != "" {
		responderMensagem(w, http.StatusOK, "se a conta estiver pendente, um novo código será enviado")
		return
	}
	if !services.EmailResendConfigurado() {
		http.Error(w, "recuperação por e-mail ainda não está configurada", http.StatusServiceUnavailable)
		return
	}
	codigo, err := emitirCodigoEmail(usuarioID, finalidadeVerificacao)
	if errors.Is(err, errCodigoEmEspera) {
		responderMensagem(w, http.StatusOK, "se a conta estiver pendente, um novo código será enviado")
		return
	}
	if err != nil {
		http.Error(w, "não foi possível enviar o código", http.StatusServiceUnavailable)
		return
	}
	if err := services.EnviarCodigoResend(email, codigo, finalidadeVerificacao); err != nil {
		invalidarCodigoAtual(usuarioID, finalidadeVerificacao)
		http.Error(w, "não foi possível enviar o código", http.StatusServiceUnavailable)
		return
	}
	responderMensagem(w, http.StatusOK, "se a conta estiver pendente, um novo código será enviado")
}

func Login(w http.ResponseWriter, r *http.Request) {
	var c credenciais
	if err := decodificarJSON(r, &c); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}
	c.Nome = strings.TrimSpace(c.Nome)
	if c.Nome == "" {
		http.Error(w, "usuário ou senha inválidos", http.StatusUnauthorized)
		return
	}

	var usuario models.Usuario
	var email, emailVerificadoEm string
	row := database.DB.QueryRow(
		"SELECT id, nome, senha_hash, token_versao, email, email_verificado_em FROM usuarios WHERE nome = ?", c.Nome,
	)
	if err := row.Scan(&usuario.ID, &usuario.Nome, &usuario.SenhaHash, &usuario.TokenVersao, &email, &emailVerificadoEm); err != nil {
		http.Error(w, "usuário ou senha inválidos", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(c.Senha)); err != nil {
		http.Error(w, "usuário ou senha inválidos", http.StatusUnauthorized)
		return
	}
	// Contas antigas ainda podem entrar para configurar o novo e-mail. Já
	// cadastros que informaram um e-mail precisam confirmá-lo antes do login.
	if email != "" && emailVerificadoEm == "" {
		http.Error(w, "confirme seu e-mail antes de entrar", http.StatusForbidden)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usuario_id":   usuario.ID,
		"nome":         usuario.Nome,
		"token_versao": usuario.TokenVersao,
		"iss":          config.EmissorJWT,
		"aud":          config.AudienciaJWT,
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenAssinado, err := token.SignedString(config.ChaveSecreta())
	if err != nil {
		http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     nomeCookieSessao,
		Value:    tokenAssinado,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSeguro(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   72 * 60 * 60,
	})
	_ = json.NewEncoder(w).Encode(map[string]string{"nome": usuario.Nome})
}

// SessaoAtual permite ao frontend restaurar a sessão sem receber o JWT no
// JavaScript. O middleware já validou o cookie antes de chegar aqui.
func SessaoAtual(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)
	var nome string
	if err := database.DB.QueryRow("SELECT nome FROM usuarios WHERE id = ?", usuarioID).Scan(&nome); err != nil {
		http.Error(w, "sessão inválida", http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"nome": nome})
}

func EncerrarSessao(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)
	if _, err := database.DB.Exec(
		"UPDATE usuarios SET token_versao = token_versao + 1 WHERE id = ?", usuarioID,
	); err != nil {
		http.Error(w, "erro ao encerrar sessão", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     nomeCookieSessao,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSeguro(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func SolicitarRecuperacaoSenha(w http.ResponseWriter, r *http.Request) {
	var entrada emailEntrada
	if err := decodificarJSON(r, &entrada); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}
	entrada.Email = normalizarEmail(entrada.Email)
	if err := validarEmail(entrada.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !services.EmailResendConfigurado() {
		http.Error(w, "recuperação por e-mail ainda não está configurada", http.StatusServiceUnavailable)
		return
	}

	var usuarioID int
	var verificadoEm string
	err := database.DB.QueryRow(
		"SELECT id, email_verificado_em FROM usuarios WHERE email = ?", entrada.Email,
	).Scan(&usuarioID, &verificadoEm)
	// A mensagem é a mesma para conta inexistente, sem e-mail ou e-mail não
	// confirmado, evitando revelar quais contas existem.
	if err != nil || verificadoEm == "" {
		responderMensagem(w, http.StatusOK, "se o e-mail estiver cadastrado, você receberá um código")
		return
	}

	codigo, err := emitirCodigoEmail(usuarioID, finalidadeRecuperacao)
	if errors.Is(err, errCodigoEmEspera) {
		responderMensagem(w, http.StatusOK, "se o e-mail estiver cadastrado, você receberá um código")
		return
	}
	if err != nil {
		http.Error(w, "não foi possível preparar o código", http.StatusInternalServerError)
		return
	}
	if err := services.EnviarCodigoResend(entrada.Email, codigo, finalidadeRecuperacao); err != nil {
		invalidarCodigoAtual(usuarioID, finalidadeRecuperacao)
		http.Error(w, "não foi possível enviar o código", http.StatusServiceUnavailable)
		return
	}
	responderMensagem(w, http.StatusOK, "se o e-mail estiver cadastrado, você receberá um código")
}

func RedefinirSenha(w http.ResponseWriter, r *http.Request) {
	var e redefinicaoEntrada
	if err := decodificarJSON(r, &e); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}
	e.Email = normalizarEmail(e.Email)
	e.Codigo = strings.TrimSpace(e.Codigo)
	if err := validarEmail(e.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validarCodigo(e.Codigo); err != nil {
		http.Error(w, "código inválido", http.StatusBadRequest)
		return
	}
	if err := validarSenha(e.NovaSenha); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var usuarioID int
	var verificadoEm string
	if err := database.DB.QueryRow(
		"SELECT id, email_verificado_em FROM usuarios WHERE email = ?", e.Email,
	).Scan(&usuarioID, &verificadoEm); err != nil || verificadoEm == "" {
		http.Error(w, "código inválido ou expirado", http.StatusUnauthorized)
		return
	}
	if err := consumirCodigoEmail(usuarioID, finalidadeRecuperacao, e.Codigo); err != nil {
		http.Error(w, "código inválido ou expirado", http.StatusUnauthorized)
		return
	}

	novoHashSenha, err := bcrypt.GenerateFromPassword([]byte(e.NovaSenha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar nova senha", http.StatusInternalServerError)
		return
	}
	if _, err := database.DB.Exec(
		"UPDATE usuarios SET senha_hash = ?, token_versao = token_versao + 1 WHERE id = ?",
		string(novoHashSenha), usuarioID,
	); err != nil {
		http.Error(w, "erro ao atualizar senha", http.StatusInternalServerError)
		return
	}
	responderMensagem(w, http.StatusOK, "senha redefinida com sucesso")
}

func ObterEmailConta(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)
	var email, verificadoEm string
	if err := database.DB.QueryRow(
		"SELECT email, email_verificado_em FROM usuarios WHERE id = ?", usuarioID,
	).Scan(&email, &verificadoEm); err != nil {
		http.Error(w, "conta não encontrada", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"email":      email,
		"verificado": verificadoEm != "",
	})
}

func SolicitarEmailConta(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)
	var entrada emailEntrada
	if err := decodificarJSON(r, &entrada); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}
	entrada.Email = normalizarEmail(entrada.Email)
	if err := validarEmail(entrada.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !services.EmailResendConfigurado() {
		http.Error(w, "recuperação por e-mail ainda não está configurada", http.StatusServiceUnavailable)
		return
	}

	var emailAnterior, verificadoAnterior string
	if err := database.DB.QueryRow(
		"SELECT email, email_verificado_em FROM usuarios WHERE id = ?", usuarioID,
	).Scan(&emailAnterior, &verificadoAnterior); err != nil {
		http.Error(w, "conta não encontrada", http.StatusNotFound)
		return
	}
	if emailAnterior == entrada.Email && verificadoAnterior != "" {
		responderMensagem(w, http.StatusOK, "este e-mail já está verificado")
		return
	}

	if _, err := database.DB.Exec(
		"UPDATE usuarios SET email = ?, email_verificado_em = '' WHERE id = ?",
		entrada.Email, usuarioID,
	); err != nil {
		http.Error(w, "este e-mail já pode estar em uso", http.StatusConflict)
		return
	}
	if emailAnterior != entrada.Email {
		// O código anterior pode ter sido enviado para outro endereço.
		// Invalidá-lo antes de emitir o novo evita vinculá-lo ao e-mail recém-informado.
		invalidarCodigoAtual(usuarioID, finalidadeVerificacao)
	}
	codigo, err := emitirCodigoEmail(usuarioID, finalidadeVerificacao)
	if err != nil && !errors.Is(err, errCodigoEmEspera) {
		_, _ = database.DB.Exec("UPDATE usuarios SET email = ?, email_verificado_em = ? WHERE id = ?", emailAnterior, verificadoAnterior, usuarioID)
		http.Error(w, "não foi possível preparar o código", http.StatusInternalServerError)
		return
	}
	if errors.Is(err, errCodigoEmEspera) {
		responderMensagem(w, http.StatusOK, "um código recente ainda está válido")
		return
	}
	if err := services.EnviarCodigoResend(entrada.Email, codigo, finalidadeVerificacao); err != nil {
		invalidarCodigoAtual(usuarioID, finalidadeVerificacao)
		_, _ = database.DB.Exec("UPDATE usuarios SET email = ?, email_verificado_em = ? WHERE id = ?", emailAnterior, verificadoAnterior, usuarioID)
		http.Error(w, "não foi possível enviar o código", http.StatusServiceUnavailable)
		return
	}
	responderMensagem(w, http.StatusOK, "código enviado para o e-mail informado")
}

func VerificarEmailConta(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)
	var entrada struct {
		Codigo string `json:"codigo"`
	}
	if err := decodificarJSON(r, &entrada); err != nil || validarCodigo(strings.TrimSpace(entrada.Codigo)) != nil {
		http.Error(w, "código inválido", http.StatusBadRequest)
		return
	}
	if err := consumirCodigoEmail(usuarioID, finalidadeVerificacao, strings.TrimSpace(entrada.Codigo)); err != nil {
		http.Error(w, "código inválido ou expirado", http.StatusUnauthorized)
		return
	}
	if _, err := database.DB.Exec(
		"UPDATE usuarios SET email_verificado_em = ? WHERE id = ?",
		time.Now().UTC().Format(time.RFC3339Nano), usuarioID,
	); err != nil {
		http.Error(w, "erro ao confirmar e-mail", http.StatusInternalServerError)
		return
	}
	responderMensagem(w, http.StatusOK, "e-mail confirmado")
}
