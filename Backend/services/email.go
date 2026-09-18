package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
	"time"
)

const endpointResend = "https://api.resend.com/emails"

type emailResend struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// EmailResendConfigurado informa se o backend pode enviar mensagens sem
// expor o conteúdo da chave em logs ou respostas HTTP.
func EmailResendConfigurado() bool {
	return strings.TrimSpace(os.Getenv("RESEND_API_KEY")) != "" &&
		strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")) != ""
}

// EnviarCodigoResend envia somente o código de uso único. A resposta do
// provedor nunca é devolvida ao cliente, pois pode conter detalhes internos.
func EnviarCodigoResend(destinatario, codigo, finalidade string) error {
	chave := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	remetente := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))
	if chave == "" || remetente == "" {
		return errors.New("serviço de e-mail não configurado")
	}

	assunto := "Confirme seu e-mail — Controle de Estoque"
	titulo := "Confirmação de e-mail"
	if finalidade == "recuperacao" {
		assunto = "Código para redefinir sua senha — Controle de Estoque"
		titulo = "Redefinição de senha"
	}

	codigoSeguro := html.EscapeString(codigo)
	corpo := emailResend{
		From:    remetente,
		To:      []string{destinatario},
		Subject: assunto,
		HTML: fmt.Sprintf(`<!doctype html><html lang="pt-BR"><body>
<h2>%s</h2>
<p>Seu código é:</p><p style="font-size:24px;font-weight:bold;letter-spacing:4px">%s</p>
<p>Ele expira em 15 minutos e só pode ser usado uma vez.</p>
<p>Se você não solicitou esta ação, ignore esta mensagem.</p>
</body></html>`, titulo, codigoSeguro),
	}

	payload, err := json.Marshal(corpo)
	if err != nil {
		return fmt.Errorf("preparar e-mail: %w", err)
	}

	requisicao, err := http.NewRequest(http.MethodPost, endpointResend, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("preparar requisição de e-mail: %w", err)
	}
	requisicao.Header.Set("Authorization", "Bearer "+chave)
	requisicao.Header.Set("Content-Type", "application/json")

	cliente := &http.Client{Timeout: 15 * time.Second}
	resposta, err := cliente.Do(requisicao)
	if err != nil {
		return fmt.Errorf("enviar e-mail: %w", err)
	}
	defer resposta.Body.Close()

	if resposta.StatusCode < http.StatusOK || resposta.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("provedor de e-mail respondeu com status %d", resposta.StatusCode)
	}
	return nil
}
