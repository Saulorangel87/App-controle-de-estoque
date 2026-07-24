package models

type Usuario struct {
	ID                    int    `json:"id"`
	Nome                  string `json:"nome"`
	SenhaHash             string `json:"-"`
	PerguntaSeguranca     string `json:"pergunta_seguranca"`
	RespostaSegurancaHash string `json:"-"`
}
