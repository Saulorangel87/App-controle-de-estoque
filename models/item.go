package models

type Item struct {
	ID            int     `json:"id"`
	UsuarioID     int     `json:"usuario_id"`
	Nome          string  `json:"nome"`
	Quantidade    float64 `json:"quantidade"`
	Unidade       string  `json:"unidade"`
	Local         string  `json:"local"`
	EstoqueMinimo float64 `json:"estoque_minimo"`
}
