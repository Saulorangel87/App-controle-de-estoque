package services

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"controle-estoque/models"
)

// Erros que o handler usa para decidir a mensagem devolvida ao frontend.
var (
	ErrOCRFalhou = errors.New(
		"não foi possível ler o texto dessa imagem — tente um print mais nítido, ou envie o XML da nota",
	)
	ErrOCRSemItens = errors.New(
		"não conseguimos identificar os itens nessa imagem — confirme se ela mostra a lista de produtos da nota",
	)
)

// ExtrairProdutosDeImagem recebe os bytes de uma foto ou print da nota
// fiscal, roda OCR (Tesseract, em português) pra extrair o texto, e
// interpreta esse texto procurando o padrão usado pela tela de consulta
// pública da SEFAZ (nome + código do produto, seguido de quantidade +
// unidade + valor unitário).
//
// O valor TOTAL de cada item não é lido da imagem — é CALCULADO
// (quantidade × valor unitário) pelo restante do fluxo, que já faz essa
// conta. Isso evita depender do OCR acertar mais um número por item, já
// que o valor total geralmente aparece separado visualmente (alinhado à
// direita), o que é mais fácil de o OCR embaralhar com outro item.
//
// FUNCIONA MELHOR com um print da tela de confirmação da SEFAZ (texto
// digital nítido) do que com foto do cupom de papel térmico (mais difícil
// pra qualquer OCR, por causa do desbotamento/reflexo/amassado do papel).
func ExtrairProdutosDeImagem(imagem []byte) ([]models.ProdXML, error) {
	texto, err := rodarOCR(imagem)
	if err != nil {
		return nil, err
	}

	produtos := interpretarTextoOCR(texto)
	if len(produtos) == 0 {
		return nil, ErrOCRSemItens
	}

	return produtos, nil
}

// rodarOCR salva a imagem num arquivo temporário e chama o Tesseract (linha
// de comando) pra extrair o texto. Usamos o binário via exec.Command, e não
// uma lib com bindings C (tipo gosseract) — assim o Dockerfile só precisa
// instalar o pacote tesseract-ocr, sem complicar a build do Go com CGO.
func rodarOCR(imagem []byte) (string, error) {
	arquivoTemp, err := os.CreateTemp("", "nota-ocr-*")
	if err != nil {
		return "", ErrOCRFalhou
	}
	defer os.Remove(arquivoTemp.Name())

	if _, err := arquivoTemp.Write(imagem); err != nil {
		arquivoTemp.Close()
		return "", ErrOCRFalhou
	}
	arquivoTemp.Close()

	// "-l por": usa o idioma português (precisa do pacote de idioma
	// instalado junto com o tesseract-ocr, ver Backend.Dockerfile).
	// "--psm 6": trata a imagem como um bloco uniforme de texto — funciona
	// melhor pra telas/listas como essa do que o modo automático padrão,
	// que é pensado pra páginas de documento com parágrafos "de verdade".
	// "stdout": manda o texto reconhecido direto pra saída padrão, sem
	// precisar gerenciar mais um arquivo temporário de resultado.
	saida, err := exec.Command(
		"tesseract", arquivoTemp.Name(), "stdout", "-l", "por", "--psm", "6",
	).Output()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOCRFalhou, err)
	}

	return string(saida), nil
}

// padraoNomeCodigo casa linhas como "LEITE INTEGRAL GODAM (Código: 22080 )".
// (?m) trata cada linha do texto separadamente. Aceita pequenas variações
// que o OCR costuma introduzir na palavra "Código" (0 no lugar do O, etc.)
// e na pontuação ao redor do número.
var padraoNomeCodigo = regexp.MustCompile(`(?m)^(.+?)\s*\(C[oó0]digo:?\s*(\d+)\s*\)`)

// padraoQuantidade casa trechos como "Qtde.:1 UN: UN Vl. Unit.: 5,89" — os
// rótulos ("Qtde", "UN", "Vl. Unit") são fixos no template da SEFAZ, então
// são um ponto de ancoragem confiável mesmo se a pontuação ao redor variar
// um pouco por causa de ruído do OCR.
var padraoQuantidade = regexp.MustCompile(
	`(?mi)Qtde\.?:?\s*([\d.,]+)\s+UN:?\s*(\S+)\s+Vl\.?\s*Unit\.?:?\s*([\d.,]+)`,
)

// interpretarTextoOCR casa cada "nome + código" (na ordem em que aparecem
// no texto) com a "quantidade + unidade + valor" correspondente (também na
// ordem em que aparecem) — o template da SEFAZ sempre intercala essas duas
// linhas por item, então parear PELA ORDEM de aparição é mais confiável do
// que tentar casar tudo numa regex só, já que o OCR quebra linha de forma
// imprevisível dependendo da imagem.
func interpretarTextoOCR(texto string) []models.ProdXML {
	nomes := padraoNomeCodigo.FindAllStringSubmatch(texto, -1)
	quantidades := padraoQuantidade.FindAllStringSubmatch(texto, -1)

	// Se as contagens não baterem, o OCR deve ter engolido ou duplicado
	// alguma linha — melhor devolver nada (o handler cai no erro
	// ErrOCRSemItens, pedindo uma imagem mais nítida) do que arriscar
	// parear um item com a quantidade errada silenciosamente.
	if len(nomes) == 0 || len(nomes) != len(quantidades) {
		return nil
	}

	produtos := make([]models.ProdXML, 0, len(nomes))
	for i := range nomes {
		nome := strings.TrimSpace(nomes[i][1])
		if nome == "" {
			continue
		}

		produtos = append(produtos, models.ProdXML{
			Nome:       nome,
			Quantidade: paraNumeroDecimal(strings.TrimSpace(quantidades[i][1])),
			Unidade:    normalizarUnidade(strings.TrimSpace(quantidades[i][2])),
			ValorUnit:  paraNumeroDecimal(strings.TrimSpace(quantidades[i][3])),
		})
	}

	return produtos
}
