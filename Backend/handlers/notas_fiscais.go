package handlers

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"

	"controle-estoque/database"
	"controle-estoque/middleware"
	"controle-estoque/models"
)

// tamanhoMaximoUpload limita o upload a 10MB — um XML de nota fiscal real
// raramente passa de algumas centenas de KB, então isso é generoso o
// suficiente sem deixar alguém mandar um arquivo gigante por engano/abuso.
const tamanhoMaximoUpload = 10 << 20 // 10MB

// ImportarNotaFiscal recebe o XML da NF-e, extrai os itens e devolve uma
// prévia comparando cada item da nota com o estoque atual do usuário — nada
// é salvo aqui ainda, só a análise. O usuário revisa no frontend e confirma
// depois via ConfirmarImportacao.
func ImportarNotaFiscal(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)

	if err := r.ParseMultipartForm(tamanhoMaximoUpload); err != nil {
		http.Error(w, "arquivo inválido ou muito grande (máximo 10MB)", http.StatusBadRequest)
		return
	}

	arquivo, _, err := r.FormFile("arquivo")
	if err != nil {
		http.Error(w, "envie o arquivo XML da nota fiscal no campo 'arquivo'", http.StatusBadRequest)
		return
	}
	defer arquivo.Close()

	conteudo, err := io.ReadAll(arquivo)
	if err != nil {
		http.Error(w, "erro ao ler o arquivo", http.StatusInternalServerError)
		return
	}

	produtos, err := extrairProdutosDoXML(conteudo)
	if err != nil {
		http.Error(w, "não foi possível interpretar o XML — confirme que é o arquivo da NF-e", http.StatusBadRequest)
		return
	}
	if len(produtos) == 0 {
		http.Error(w, "nenhum item encontrado nesse XML", http.StatusBadRequest)
		return
	}

	itensEstoque, err := buscarItensDoUsuario(usuarioID)
	if err != nil {
		http.Error(w, "erro ao consultar o estoque atual", http.StatusInternalServerError)
		return
	}

	resultado := make([]models.ItemImportado, 0, len(produtos))
	for _, produto := range produtos {
		quantidade, _ := strconv.ParseFloat(strings.TrimSpace(produto.Quantidade), 64)

		item := models.ItemImportado{
			NomeNota:       produto.Nome,
			QuantidadeNota: quantidade,
			UnidadeNota:    produto.Unidade,
			Status:         "novo",
		}

		if encontrado, ok := encontrarCorrespondencia(produto.Nome, itensEstoque); ok {
			item.Status = "encontrado"
			item.ItemID = &encontrado.ID
			item.NomeAtual = encontrado.Nome
			item.QuantidadeAtual = encontrado.Quantidade
		}

		resultado = append(resultado, item)
	}

	json.NewEncoder(w).Encode(resultado)
}

// ConfirmarImportacao recebe a lista já revisada pelo usuário (com os itens
// "não identificados" resolvidos manualmente) e aplica de fato: soma
// quantidade nos itens encontrados, cria os itens novos confirmados.
func ConfirmarImportacao(w http.ResponseWriter, r *http.Request) {
	usuarioID := r.Context().Value(middleware.UsuarioIDContexto).(int)

	var entradas []models.EntradaConfirmada
	if err := json.NewDecoder(r.Body).Decode(&entradas); err != nil {
		http.Error(w, "dados inválidos", http.StatusBadRequest)
		return
	}

	atualizados := 0
	criados := 0

	for _, entrada := range entradas {
		if entrada.ItemID != nil {
			resultado, err := database.DB.Exec(
				"UPDATE itens SET quantidade = quantidade + ? WHERE id = ? AND usuario_id = ?",
				entrada.Quantidade, *entrada.ItemID, usuarioID,
			)
			if err != nil {
				continue
			}
			if linhas, _ := resultado.RowsAffected(); linhas > 0 {
				atualizados++
			}
			continue
		}

		if entrada.Nome == "" || entrada.Unidade == "" || entrada.Local == "" {
			continue
		}

		_, err := database.DB.Exec(
			`INSERT INTO itens (usuario_id, nome, quantidade, unidade, local, estoque_minimo)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			usuarioID, entrada.Nome, entrada.Quantidade, entrada.Unidade, entrada.Local, entrada.EstoqueMinimo,
		)
		if err == nil {
			criados++
		}
	}

	json.NewEncoder(w).Encode(map[string]int{
		"atualizados": atualizados,
		"criados":     criados,
	})
}

// --- funções auxiliares ---

// extrairProdutosDoXML tenta os dois formatos possíveis do XML da NF-e:
// com envelope <nfeProc> (mais comum ao baixar do portal da SEFAZ) ou só
// <NFe> direto (comum quando o próprio emissor exporta a nota).
func extrairProdutosDoXML(conteudo []byte) ([]models.ProdXML, error) {
	var comEnvelope models.NfeProcXML
	if err := xml.Unmarshal(conteudo, &comEnvelope); err == nil && len(comEnvelope.NFe.InfNFe.Itens) > 0 {
		return coletarProdutos(comEnvelope.NFe.InfNFe.Itens), nil
	}

	var semEnvelope models.NFeXML
	if err := xml.Unmarshal(conteudo, &semEnvelope); err != nil {
		return nil, err
	}
	return coletarProdutos(semEnvelope.InfNFe.Itens), nil
}

func coletarProdutos(itens []models.DetXML) []models.ProdXML {
	produtos := make([]models.ProdXML, 0, len(itens))
	for _, item := range itens {
		produtos = append(produtos, item.Prod)
	}
	return produtos
}

type itemEstoqueSimples struct {
	ID         int
	Nome       string
	Quantidade float64
}

func buscarItensDoUsuario(usuarioID int) ([]itemEstoqueSimples, error) {
	rows, err := database.DB.Query(
		"SELECT id, nome, quantidade FROM itens WHERE usuario_id = ?", usuarioID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var itens []itemEstoqueSimples
	for rows.Next() {
		var item itemEstoqueSimples
		if err := rows.Scan(&item.ID, &item.Nome, &item.Quantidade); err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}
	return itens, nil
}

// encontrarCorrespondencia compara o nome do produto da nota com os nomes já
// cadastrados no estoque. Primeiro tenta igualdade exata (ignorando
// maiúsculas/minúsculas e espaços), depois uma comparação mais flexível por
// substring (pega casos como "Peito de Frango Resfriado" da nota batendo com
// "Peito de Frango" já cadastrado). É uma heurística simples — itens com
// nomes muito diferentes do que está na nota continuam caindo em "novo",
// exigindo associação manual na tela de conferência.
func encontrarCorrespondencia(nomeProduto string, itens []itemEstoqueSimples) (itemEstoqueSimples, bool) {
	alvo := normalizarNome(nomeProduto)

	for _, item := range itens {
		if normalizarNome(item.Nome) == alvo {
			return item, true
		}
	}

	for _, item := range itens {
		nomeItem := normalizarNome(item.Nome)
		if strings.Contains(alvo, nomeItem) || strings.Contains(nomeItem, alvo) {
			return item, true
		}
	}

	return itemEstoqueSimples{}, false
}

func normalizarNome(nome string) string {
	return strings.ToLower(strings.TrimSpace(nome))
}
