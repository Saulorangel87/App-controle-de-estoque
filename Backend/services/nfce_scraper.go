package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"controle-estoque/models"
)

// dominiosPermitidosNFCe restringe a consulta apenas a domínios oficiais de
// SEFAZ que já validamos. Isso evita que o backend seja usado como um proxy
// genérico para buscar qualquer URL que alguém mandar (risco de SSRF).
//
// ESTADO ATUAL: NENHUM domínio habilitado. Confirmado, testando com uma nota
// real do RJ, que a consulta pública de NFC-e da SEFAZ-RJ
// (consultadfe.fazenda.rj.gov.br) exige reCAPTCHA v3 (grecaptcha.execute) e
// tem proteção anti-bot corporativa (scripts "/TSPD/", característicos do
// F5 Shape / Distributed Cloud Bot Defense) — a página só entrega os dados
// depois de um POST AJAX autenticado por esse token. Isso não é contornável
// de forma confiável (nem correto tentar) via scraping server-side, então a
// via fica DESATIVADA por enquanto — a mensagem em ErrDominioNaoSuportado
// abaixo é a que o usuário vê ao tentar. Se algum dia outro estado (UF) for
// avaliado e não tiver esse tipo de proteção, é aqui que ele entraria.
var dominiosPermitidosNFCe = map[string]bool{}

// Erros que o handler usa para decidir o status HTTP e a mensagem devolvida
// ao frontend.
var (
	ErrURLInvalida          = errors.New("link inválido — cole a URL completa lida no QR Code da nota")
	ErrDominioNaoSuportado  = errors.New("a importação por QR Code está indisponível no momento — o site da SEFAZ exige verificação anti-robô nessa consulta. Envie o XML da nota em vez disso")
	ErrConsultaFalhou       = errors.New("não foi possível consultar essa nota na SEFAZ agora — tente novamente em instantes ou envie o XML")
	ErrNenhumItemEncontrado = errors.New("não conseguimos identificar os itens dessa nota — confirme se o link do QR Code está correto")
)

// ConsultarProdutosPorURL recebe a URL extraída do QR Code de uma NFC-e,
// consulta a página pública da SEFAZ e devolve os produtos no MESMO formato
// (models.ProdXML) que o parser de XML já usa em handlers/notas_fiscais.go —
// assim o resto do fluxo de importação (comparar com o estoque, criar ou
// somar quantidade) é 100% reaproveitado, sem duplicar lógica de negócio.
//
// ATENÇÃO — leia antes de mexer aqui:
// Isso é web scraping de uma página HTML pública, não uma API oficial da
// SEFAZ (ela não existe). O layout da página de consulta pode mudar sem
// aviso prévio, o que quebraria o parser abaixo. Por isso a extração usa uma
// estratégia em duas camadas (ver extrairProdutosDoHTML) em vez de depender
// de uma classe/id específico só — mas se a SEFAZ mudar o site e isso parar
// de funcionar, o jeito de corrigir é: abrir a nota real no navegador,
// inspecionar o HTML da tabela de itens e ajustar os seletores/heurísticas
// abaixo de acordo.
func ConsultarProdutosPorURL(urlNota string) ([]models.ProdXML, error) {
	urlValidada, err := validarURL(urlNota)
	if err != nil {
		return nil, err
	}

	documento, err := buscarDocumento(urlValidada)
	if err != nil {
		return nil, err
	}

	produtos := extrairProdutosDoHTML(documento)
	if len(produtos) == 0 {
		return nil, ErrNenhumItemEncontrado
	}

	return produtos, nil
}

// validarURL confirma que a URL é HTTPS e pertence a um domínio de SEFAZ
// que já suportamos, antes de fazer qualquer requisição.
func validarURL(bruta string) (string, error) {
	bruta = strings.TrimSpace(bruta)

	analisada, err := url.Parse(bruta)
	if err != nil || analisada.Host == "" || (analisada.Scheme != "https" && analisada.Scheme != "http") {
		return "", ErrURLInvalida
	}

	if !dominiosPermitidosNFCe[analisada.Host] {
		return "", ErrDominioNaoSuportado
	}

	return analisada.String(), nil
}

// buscarDocumento faz a requisição HTTP à SEFAZ com um timeout curto e um
// User-Agent de navegador comum (algumas SEFAZ bloqueiam requisições sem
// esse cabeçalho por parecerem bots muito simples) e devolve o HTML já
// carregado no goquery para facilitar a extração.
func buscarDocumento(urlNota string) (*goquery.Document, error) {
	cliente := &http.Client{Timeout: 20 * time.Second}

	requisicao, err := http.NewRequest(http.MethodGet, urlNota, nil)
	if err != nil {
		return nil, ErrConsultaFalhou
	}
	requisicao.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36",
	)
	requisicao.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

	resposta, err := cliente.Do(requisicao)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConsultaFalhou, err)
	}
	defer resposta.Body.Close()

	if resposta.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w (status %d)", ErrConsultaFalhou, resposta.StatusCode)
	}

	corpoBruto, err := io.ReadAll(resposta.Body)
	if err != nil {
		return nil, ErrConsultaFalhou
	}

	// DEBUG TEMPORÁRIO — enquanto o parser ainda não foi validado contra o
	// site real: salva sempre o HTML exatamente como o Go recebeu (pode ser
	// diferente do que aparece no navegador, se a página carregar itens via
	// JavaScript depois do carregamento inicial). Depois que o parser
	// estiver validado, pode remover este bloco e a função salvarHTMLDebug.
	salvarHTMLDebug(corpoBruto)

	documento, err := goquery.NewDocumentFromReader(bytes.NewReader(corpoBruto))
	if err != nil {
		return nil, ErrConsultaFalhou
	}

	return documento, nil
}

// salvarHTMLDebug grava a última resposta da SEFAZ num arquivo local, na
// pasta onde o backend está rodando. É só uma ferramenta de depuração
// enquanto ajustamos o parser — não afeta o funcionamento normal do app
// (falha em salvar é ignorada de propósito, nunca deve quebrar a consulta).
func salvarHTMLDebug(html []byte) {
	_ = os.WriteFile("debug_nfce_ultima_consulta.html", html, 0o644)
}

// padrãoValorMonetario reconhece números no formato brasileiro (vírgula
// decimal, com ou sem separador de milhar em ponto) — ex: "12,50", "1.234,56".
var padraoValorMonetario = regexp.MustCompile(`^\d{1,3}(\.\d{3})*,\d{2,4}$|^\d+,\d{2,4}$`)

// unidadesConhecidas ajuda a identificar, entre as células de uma linha,
// qual delas é a unidade de medida (em vez de nome do produto ou valor).
var unidadesConhecidas = regexp.MustCompile(`(?i)^(UN|UND|UNID|KG|G|GR|L|ML|CX|PC|PCT|DZ|PAR|FD|SC)\.?$`)

// extrairProdutosDoHTML procura, em qualquer tabela da página, linhas que
// pareçam itens de compra: uma célula com o nome do produto, uma com
// quantidade, uma com unidade de medida e uma com valor. A checagem não
// depende de um id/classe específico da SEFAZ — em vez disso, avalia o
// FORMATO de cada célula da linha, o que tende a ser mais resistente a
// pequenas mudanças de layout (mas não a garantia absoluta — ver o
// comentário no topo do arquivo).
func extrairProdutosDoHTML(documento *goquery.Document) []models.ProdXML {
	var produtos []models.ProdXML

	documento.Find("tr").Each(func(_ int, linha *goquery.Selection) {
		celulas := linha.Find("td")
		if celulas.Length() < 3 {
			return // cabeçalho ou linha decorativa, não item de compra
		}

		textos := make([]string, 0, celulas.Length())
		celulas.Each(func(_ int, celula *goquery.Selection) {
			textos = append(textos, strings.TrimSpace(celula.Text()))
		})

		produto, encontrado := interpretarLinha(textos)
		if encontrado {
			produtos = append(produtos, produto)
		}
	})

	return produtos
}

// interpretarLinha tenta montar um ProdXML a partir dos textos de uma linha
// de tabela, identificando cada campo pelo formato do conteúdo em vez da
// posição fixa da coluna (colunas variam de portal para portal).
func interpretarLinha(celulas []string) (models.ProdXML, bool) {
	var produto models.ProdXML
	var quantidadeBruta, valorBruto string
	nomeCandidato := ""

	for _, texto := range celulas {
		semEspacos := strings.TrimSpace(texto)
		if semEspacos == "" {
			continue
		}

		switch {
		case unidadesConhecidas.MatchString(semEspacos) && produto.Unidade == "":
			produto.Unidade = normalizarUnidade(semEspacos)

		case padraoValorMonetario.MatchString(semEspacos):
			// A primeira ocorrência nesse formato costuma ser a quantidade
			// (ex: "2,000"); a(s) seguinte(s) tendem a ser valores em R$.
			// Guardamos as duas primeiras e decidimos depois pelo contexto.
			if quantidadeBruta == "" {
				quantidadeBruta = semEspacos
			} else if valorBruto == "" {
				valorBruto = semEspacos
			}

		default:
			// Texto "comum": candidato a nome do produto — fica com o mais
			// longo encontrado na linha, que costuma ser a descrição.
			if len(semEspacos) > len(nomeCandidato) {
				nomeCandidato = semEspacos
			}
		}
	}

	// Uma linha só conta como item de compra se tivermos, no mínimo, nome +
	// quantidade + unidade. Sem isso, é provável que seja uma linha de
	// totais, tributos ou rodapé — que devem ser ignoradas.
	if nomeCandidato == "" || quantidadeBruta == "" || produto.Unidade == "" {
		return models.ProdXML{}, false
	}

	produto.Nome = nomeCandidato
	produto.Quantidade = paraNumeroDecimal(quantidadeBruta)
	produto.ValorUnit = paraNumeroDecimal(valorBruto)

	return produto, true
}

// paraNumeroDecimal converte um número no formato brasileiro ("1.234,56" ou
// "2,000") para o formato com ponto decimal ("1234.56" / "2.000") que o
// resto do código (herdado do parser de XML) já espera via strconv.ParseFloat.
func paraNumeroDecimal(brasileiro string) string {
	if brasileiro == "" {
		return "0"
	}
	semMilhar := strings.ReplaceAll(brasileiro, ".", "")
	comPontoDecimal := strings.ReplaceAll(semMilhar, ",", ".")
	return comPontoDecimal
}

// normalizarUnidade agrupa variações comuns de escrita da mesma unidade de
// medida (o texto de fábrica/nota nem sempre é padronizado) na forma que o
// resto do sistema usa. Isso evita, por exemplo, ter "UN" e "UND" tratados
// como unidades diferentes na hora de comparar com o estoque já cadastrado.
func normalizarUnidade(bruta string) string {
	bruta = strings.ToUpper(strings.TrimSuffix(strings.TrimSpace(bruta), "."))

	switch bruta {
	case "UN", "UND", "UNID", "UNIDADE":
		return "UN"
	case "KG", "QUILO", "KILO":
		return "KG"
	case "G", "GR", "GRAMA":
		return "G"
	case "L", "LT", "LITRO":
		return "L"
	case "ML":
		return "ML"
	case "CX":
		return "CX"
	case "PC", "PCT", "PCTE":
		return "PC"
	case "DZ":
		return "DZ"
	case "PAR":
		return "PAR"
	case "FD", "SC":
		return bruta
	default:
		return bruta
	}
}
