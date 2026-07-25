package models

// Structs que espelham só os campos que precisamos do XML padrão da NF-e
// (Nota Fiscal Eletrônica). O XML oficial tem dezenas de campos; ignoramos
// tudo que não seja o essencial para o estoque (nome, quantidade, unidade).
//
// O XML pode vir de duas formas dependendo de onde foi baixado:
//   1) Envelope completo: <nfeProc><NFe><infNFe>...
//   2) Só a nota, sem envelope: <NFe><infNFe>...
// O parser tenta os dois formatos (ver handlers/notas_fiscais.go).

type NfeProcXML struct {
	NFe NFeXML `xml:"NFe"`
}

type NFeXML struct {
	InfNFe InfNFeXML `xml:"infNFe"`
}

type InfNFeXML struct {
	Itens []DetXML `xml:"det"`
}

type DetXML struct {
	Prod ProdXML `xml:"prod"`
}

type ProdXML struct {
	Codigo     string `xml:"cProd"`
	Nome       string `xml:"xProd"`
	Unidade    string `xml:"uCom"`
	Quantidade string `xml:"qCom"` // vem como texto (ex: "5.0000"), convertido depois
	ValorUnit  string `xml:"vUnCom"`
}

// ItemImportado é o formato "limpo" que devolvemos para o frontend depois de
// interpretar o XML e comparar com o estoque atual do usuário.
type ItemImportado struct {
	NomeNota      string  `json:"nome_nota"`
	QuantidadeNota float64 `json:"quantidade_nota"`
	UnidadeNota   string  `json:"unidade_nota"`
	Status        string  `json:"status"` // "encontrado" ou "novo"
	ItemID        *int    `json:"item_id,omitempty"`
	NomeAtual     string  `json:"nome_atual,omitempty"`
	QuantidadeAtual float64 `json:"quantidade_atual,omitempty"`
}

// EntradaConfirmada é o que o frontend manda de volta depois que o usuário
// revisou a lista (associou itens não identificados, confirmou os novos).
type EntradaConfirmada struct {
	ItemID        *int    `json:"item_id"` // null = criar item novo
	Nome          string  `json:"nome"`
	Quantidade    float64 `json:"quantidade"` // quantidade a SOMAR (não o total final)
	Unidade       string  `json:"unidade"`
	Local         string  `json:"local"`          // só usado se item_id for null
	EstoqueMinimo float64 `json:"estoque_minimo"` // só usado se item_id for null
}
