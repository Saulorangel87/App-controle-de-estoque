package services

import (
<<<<<<< HEAD
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // registra o decoder de JPEG (image.Decode detecta o formato sozinho)
	"image/png"
	"os"
	"os/exec"
	"regexp"
	"strconv"
=======
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
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
<<<<<<< HEAD
// interpreta esse texto procurando um dos formatos conhecidos (tela da
// SEFAZ ou cupom de papel impresso — ver interpretarTextoOCR).
=======
// interpreta esse texto procurando o padrão usado pela tela de consulta
// pública da SEFAZ (nome + código do produto, seguido de quantidade +
// unidade + valor unitário).
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
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
<<<<<<< HEAD
	imagemDecodificada, _, err := image.Decode(bytes.NewReader(imagem))
	if err != nil {
		// Formato de imagem não suportado ou arquivo corrompido — não tem
		// como rotacionar/pré-processar nesse caso, tenta OCR direto nos
		// bytes originais como última tentativa.
		texto, errOCR := rodarOCR(imagem)
		if errOCR != nil {
			return nil, errOCR
		}
		return finalizarExtracao(texto, nil)
	}

	// Foto tirada com celular pode vir em qualquer orientação — o
	// Tesseract não gira a imagem sozinho, e texto "deitado" praticamente
	// não é reconhecido. Testamos as 4 rotações possíveis (0°, 90°, 180°,
	// 270°) e ficamos com a que reconhecer mais itens. Roda um pouco mais
	// lento (até 4 chamadas ao Tesseract em vez de 1), mas é uma operação
	// sob demanda, não algo que precise ser instantâneo.
	var melhorTexto string
	var melhoresProdutos []models.ProdXML
	var ultimoErro error
	primeiraTentativa := true

	for rotacoes := 0; rotacoes < 4; rotacoes++ {
		imagemRotacionada := rotacionar90(imagemDecodificada, rotacoes)

		imagemProcessada, err := prepararImagemParaOCR(imagemRotacionada)
		if err != nil {
			ultimoErro = err
			continue
		}

		texto, err := rodarOCR(imagemProcessada)
		if err != nil {
			ultimoErro = err
			continue
		}

		// Garante que SEMPRE sobra algum texto pra debug, mesmo que
		// nenhuma rotação encontre itens — sem isso, quando as 4
		// tentativas falham, melhorTexto ficaria vazio (string zero-value)
		// em vez de mostrar o que o Tesseract realmente leu.
		if primeiraTentativa {
			melhorTexto = texto
			primeiraTentativa = false
		}

		produtos := interpretarTextoOCR(texto)
		if len(produtos) > len(melhoresProdutos) {
			melhoresProdutos = produtos
			melhorTexto = texto
		}
	}

	if primeiraTentativa {
		// Chegou aqui e "primeiraTentativa" continua true quer dizer que
		// as 4 rotações falharam antes mesmo de conseguir rodar o OCR
		// (nenhuma teve chance de setar melhorTexto) — isso é diferente
		// de "rodou o OCR mas não achou os itens". Devolve o erro real do
		// Tesseract em vez de mascarar como ErrOCRSemItens, senão fica
		// impossível saber (e debugar) o que realmente deu errado.
		if ultimoErro != nil {
			return nil, ultimoErro
		}
		return nil, ErrOCRFalhou
	}

	return finalizarExtracao(melhorTexto, melhoresProdutos)
}

// finalizarExtracao centraliza a decisão final: se algum ângulo testado
// encontrou itens, devolve o melhor resultado; senão, salva o texto de
// debug (da melhor tentativa, mesmo vazia) e devolve o erro apropriado.
func finalizarExtracao(texto string, produtos []models.ProdXML) ([]models.ProdXML, error) {
	if len(produtos) == 0 {
		// DEBUG TEMPORÁRIO — enquanto o parser ainda está sendo calibrado:
		// salva o texto exatamente como o Tesseract devolveu, pra dar pra
		// ver o que ele realmente leu (em vez de advinhar pelo regex).
		// Depois que o parser estiver validado, pode remover este bloco e
		// a função salvarDebugOCR.
		salvarDebugOCR(texto)
=======
	texto, err := rodarOCR(imagem)
	if err != nil {
		return nil, err
	}

	produtos := interpretarTextoOCR(texto)
	if len(produtos) == 0 {
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
		return nil, ErrOCRSemItens
	}

	return produtos, nil
}

<<<<<<< HEAD
// salvarDebugOCR grava o texto bruto da última leitura de OCR num arquivo
// local, na pasta onde o backend está rodando. É só uma ferramenta de
// depuração enquanto ajustamos o parser — falha em salvar é ignorada de
// propósito, nunca deve quebrar a resposta ao usuário.
func salvarDebugOCR(texto string) {
	_ = os.WriteFile("debug_ocr_ultimo_texto.txt", []byte(texto), 0o644)
}

// rodarOCR salva a imagem (já pré-processada) num arquivo temporário e
// chama o Tesseract (linha de comando) pra extrair o texto. Usamos o
// binário via exec.Command, e não uma lib com bindings C (tipo gosseract)
// — assim o Dockerfile só precisa instalar o pacote tesseract-ocr, sem
// complicar a build do Go com CGO.
func rodarOCR(imagemProcessada []byte) (string, error) {
	arquivoTemp, err := os.CreateTemp("", "nota-ocr-*.png")
=======
// rodarOCR salva a imagem num arquivo temporário e chama o Tesseract (linha
// de comando) pra extrair o texto. Usamos o binário via exec.Command, e não
// uma lib com bindings C (tipo gosseract) — assim o Dockerfile só precisa
// instalar o pacote tesseract-ocr, sem complicar a build do Go com CGO.
func rodarOCR(imagem []byte) (string, error) {
	arquivoTemp, err := os.CreateTemp("", "nota-ocr-*")
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
	if err != nil {
		return "", ErrOCRFalhou
	}
	defer os.Remove(arquivoTemp.Name())

<<<<<<< HEAD
	if _, err := arquivoTemp.Write(imagemProcessada); err != nil {
=======
	if _, err := arquivoTemp.Write(imagem); err != nil {
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
		arquivoTemp.Close()
		return "", ErrOCRFalhou
	}
	arquivoTemp.Close()

	// "-l por": usa o idioma português (precisa do pacote de idioma
	// instalado junto com o tesseract-ocr, ver Backend.Dockerfile).
	// "--psm 6": trata a imagem como um bloco uniforme de texto — funciona
	// melhor pra telas/listas como essa do que o modo automático padrão,
	// que é pensado pra páginas de documento com parágrafos "de verdade".
<<<<<<< HEAD
	// "--dpi 300": screenshots normalmente não têm informação de DPI
	// embutida, e sem isso o Tesseract às vezes assume uma resolução baixa
	// e piora a leitura de números pequenos — forçar 300 evita essa
	// suposição errada.
	// "stdout": manda o texto reconhecido direto pra saída padrão, sem
	// precisar gerenciar mais um arquivo temporário de resultado.
	saida, err := exec.Command(
		"tesseract", arquivoTemp.Name(), "stdout",
		"-l", "por", "--psm", "6", "--dpi", "300",
=======
	// "stdout": manda o texto reconhecido direto pra saída padrão, sem
	// precisar gerenciar mais um arquivo temporário de resultado.
	saida, err := exec.Command(
		"tesseract", arquivoTemp.Name(), "stdout", "-l", "por", "--psm", "6",
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
	).Output()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOCRFalhou, err)
	}

	return string(saida), nil
}

<<<<<<< HEAD
// rotacionar90 gira a imagem em incrementos de 90° no sentido horário.
// "vezes" pode ser 0 (sem rotação), 1 (90°), 2 (180°) ou 3 (270°) — usado
// pra testar todas as orientações possíveis de uma foto tirada com celular
// (ver ExtrairProdutosDeImagem).
func rotacionar90(img image.Image, vezes int) image.Image {
	resultado := img
	for i := 0; i < vezes%4; i++ {
		resultado = rotacionarUmaVez(resultado)
	}
	return resultado
}

// rotacionarUmaVez gira a imagem 90° no sentido horário: o pixel que
// estava em (x, y) vai para (altura-1-y, x) na imagem girada.
func rotacionarUmaVez(img image.Image) image.Image {
	limites := img.Bounds()
	largura := limites.Dx()
	altura := limites.Dy()

	girada := image.NewRGBA(image.Rect(0, 0, altura, largura))
	for y := 0; y < altura; y++ {
		for x := 0; x < largura; x++ {
			girada.Set(altura-1-y, x, img.At(limites.Min.X+x, limites.Min.Y+y))
		}
	}
	return girada
}

// fatorAmpliacao multiplica a resolução da imagem antes do OCR. Testando
// contra uma nota real, o Tesseract estava perdendo dígitos pequenos e finos
// (principalmente na coluna de quantidade) — aumentar a imagem dá mais
// pixels por traço de cada caractere, o que costuma melhorar bastante o
// reconhecimento de texto pequeno em prints/screenshots.
const fatorAmpliacao = 3

// prepararImagemParaOCR recebe uma imagem já decodificada (e, se for o
// caso, já rotacionada — ver rotacionar90), converte pra escala de cinza e
// amplia a resolução, devolvendo um novo PNG. Escala de cinza reduz ruído
// de cor que não ajuda em nada o reconhecimento de texto; ampliar dá mais
// definição pros traços finos dos caracteres.
func prepararImagemParaOCR(imagemDecodificada image.Image) ([]byte, error) {
	limites := imagemDecodificada.Bounds()
	larguraOriginal := limites.Dx()
	alturaOriginal := limites.Dy()

	imagemAmpliada := image.NewGray(image.Rect(
		0, 0, larguraOriginal*fatorAmpliacao, alturaOriginal*fatorAmpliacao,
	))

	// Amostragem "vizinho mais próximo": cada pixel da imagem ampliada
	// copia o pixel correspondente da original — simples e suficiente aqui,
	// já que o objetivo é só dar mais espaço em pixels pros traços dos
	// caracteres, não suavizar a imagem (suavizar poderia até atrapalhar,
	// borrando bordas que o OCR usa pra distinguir letras/números).
	for y := 0; y < alturaOriginal*fatorAmpliacao; y++ {
		yOriginal := limites.Min.Y + y/fatorAmpliacao
		for x := 0; x < larguraOriginal*fatorAmpliacao; x++ {
			xOriginal := limites.Min.X + x/fatorAmpliacao
			corOriginal := imagemDecodificada.At(xOriginal, yOriginal)
			cinza := color.GrayModel.Convert(corOriginal).(color.Gray)
			imagemAmpliada.SetGray(x, y, cinza)
		}
	}

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imagemAmpliada); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

=======
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
// padraoNomeCodigo casa linhas como "LEITE INTEGRAL GODAM (Código: 22080 )".
// (?m) trata cada linha do texto separadamente. Aceita pequenas variações
// que o OCR costuma introduzir na palavra "Código" (0 no lugar do O, etc.)
// e na pontuação ao redor do número.
var padraoNomeCodigo = regexp.MustCompile(`(?m)^(.+?)\s*\(C[oó0]digo:?\s*(\d+)\s*\)`)

// padraoQuantidade casa trechos como "Qtde.:1 UN: UN Vl. Unit.: 5,89" — os
// rótulos ("Qtde", "UN", "Vl. Unit") são fixos no template da SEFAZ, então
// são um ponto de ancoragem confiável mesmo se a pontuação ao redor variar
<<<<<<< HEAD
// um pouco por causa de ruído do OCR. O [VW][l1I] aceita "Vl", "V1", "VI"
// ou "WI" — testando contra notas reais, o Tesseract já leu "Vl." tanto como
// "VI." (confunde "l" minúsculo com "I" maiúsculo) quanto como "WI." (W no
// lugar do V) — erros clássicos de OCR nessa fonte/resolução.
//
// O grupo da quantidade aceita dígitos, pontuação OU "?" ([\d.,?]*, não +):
// testando contra notas reais, o dígito da quantidade (frequentemente "1")
// às vezes desaparece completamente do texto reconhecido, e outras vezes
// vira um "?" literal — não é erro de formato, o OCR simplesmente não "viu"
// aquele número direito. Em ambos os casos, interpretarFormatoTelaSefaz
// assume quantidade 1 como padrão (ver comentário lá) — a tela de revisão
// no frontend é quem garante que um valor errado seja corrigido antes de
// confirmar, então esse chute é seguro.
var padraoQuantidade = regexp.MustCompile(
	`(?mi)Qtde\.?:?\s*([\d.,?]*)\s+UN:?\s*(\S+)\s+[VW][l1I]\.?\s*Unit\.?:?\s*([\d.,]+)`,
)

// padraoItemCupomPapel casa uma linha inteira do cupom de papel impresso,
// formato bem diferente da tela do navegador: tudo numa linha só —
// "código  descrição  quantidade  unidade  x  preço_unitário  total"
// (ex: "7096183202187 LEITE INTEGRAL QUATA 1 UN x 5,49    5,49"). O total
// no final é capturado só pra ancorar o fim da linha — não é usado (o valor
// total de cada item continua sendo calculado, não lido, pelos mesmos
// motivos explicados em ExtrairProdutosDeImagem).
var padraoItemCupomPapel = regexp.MustCompile(
	`(?m)^(\d{4,20})\s+(.+?)\s+([\d.,]+)\s+(\S{1,6})\s*[xX]\s*([\d.,]+)\s+([\d.,]+)\s*$`,
)

// padraoItemComIndicadorImposto casa um terceiro formato de cupom de papel,
// usado por outras redes: "número_do_item  código  descrição  qtd+unidade
// colados  indicador_de_imposto  preço" (ex: "001 08871124 HIDRAT JOHNSONS
// SOFT 1un F1 10,39)"). O indicador de imposto (a letra exigida pela Lei da
// Transparência Fiscal, tipo "F1" ou "T19,00%") não nos interessa — o \S+
// antes do preço só serve pra pular esse token, seja lá qual for.
//
// O grupo da quantidade (\S+?) vem ANTES de "un" colado (sem espaço) —
// testando contra uma nota real, o Tesseract leu o "1" de "1un" como "l"
// minúsculo ou "j" (mais confusão clássica de dígito/letra), então em vez
// de exigir um dígito ali, capturamos qualquer coisa e, se não der pra
// interpretar como número, assumimos 1 (mesmo raciocínio do formato da
// tela da SEFAZ — a revisão manual corrige se estiver errado).
var padraoItemComIndicadorImposto = regexp.MustCompile(
	`(?mi)^\d{3}\s+(\d+)\s+(.+?)\s+(\S+?)[uU][nN]\s+\S+\s+([\d.,]+)\)?\s*$`,
)

// interpretarTextoOCR tenta reconhecer, em ordem, os formatos conhecidos de
// nota/cupom — cada rede/portal usa um layout diferente, então em vez de
// uma tentativa só, tenta cada um até algum encontrar itens:
//  1. Tela da SEFAZ (mais comum quando a origem é um print/screenshot)
//  2. Cupom de papel com "qtd unidade x preço total"
//  3. Cupom de papel com número do item + indicador de imposto
func interpretarTextoOCR(texto string) []models.ProdXML {
	if produtos := interpretarFormatoTelaSefaz(texto); len(produtos) > 0 {
		return produtos
	}
	if produtos := interpretarFormatoCupomPapel(texto); len(produtos) > 0 {
		return produtos
	}
	return interpretarFormatoCupomComIndicadorImposto(texto)
}

// interpretarFormatoTelaSefaz casa cada "nome + código" (na ordem em que
// aparecem no texto) com a "quantidade + unidade + valor" correspondente
// (também na ordem em que aparecem) — o template da tela da SEFAZ sempre
// intercala essas duas linhas por item, então parear PELA ORDEM de
// aparição é mais confiável do que tentar casar tudo numa regex só, já que
// o OCR quebra linha de forma imprevisível dependendo da imagem.
func interpretarFormatoTelaSefaz(texto string) []models.ProdXML {
=======
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
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
	nomes := padraoNomeCodigo.FindAllStringSubmatch(texto, -1)
	quantidades := padraoQuantidade.FindAllStringSubmatch(texto, -1)

	// Se as contagens não baterem, o OCR deve ter engolido ou duplicado
<<<<<<< HEAD
	// alguma linha — melhor devolver nada (interpretarTextoOCR cai pro
	// formato de cupom de papel, e se esse também não achar nada, o
	// handler cai no erro ErrOCRSemItens) do que arriscar parear um item
	// com a quantidade errada silenciosamente.
=======
	// alguma linha — melhor devolver nada (o handler cai no erro
	// ErrOCRSemItens, pedindo uma imagem mais nítida) do que arriscar
	// parear um item com a quantidade errada silenciosamente.
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
	if len(nomes) == 0 || len(nomes) != len(quantidades) {
		return nil
	}

	produtos := make([]models.ProdXML, 0, len(nomes))
	for i := range nomes {
		nome := strings.TrimSpace(nomes[i][1])
		if nome == "" {
			continue
		}

<<<<<<< HEAD
		quantidadeLida := strings.ReplaceAll(strings.TrimSpace(quantidades[i][1]), "?", "")
		if quantidadeLida == "" {
			// O dígito sumiu do OCR — assume 1 (o mais comum em compras),
			// e conta com a revisão manual pra corrigir se estiver errado.
			quantidadeLida = "1"
		}

		produtos = append(produtos, models.ProdXML{
			Nome:       nome,
			Quantidade: paraNumeroDecimal(quantidadeLida),
=======
		produtos = append(produtos, models.ProdXML{
			Nome:       nome,
			Quantidade: paraNumeroDecimal(strings.TrimSpace(quantidades[i][1])),
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
			Unidade:    normalizarUnidade(strings.TrimSpace(quantidades[i][2])),
			ValorUnit:  paraNumeroDecimal(strings.TrimSpace(quantidades[i][3])),
		})
	}

	return produtos
}
<<<<<<< HEAD

// interpretarFormatoCupomPapel casa o formato de uma linha só usado no
// cupom de papel impresso (ver padraoItemCupomPapel). Diferente do formato
// da tela, aqui cada linha já tem tudo junto, então não precisa parear por
// ordem — cada match da regex já é um item completo.
func interpretarFormatoCupomPapel(texto string) []models.ProdXML {
	itens := padraoItemCupomPapel.FindAllStringSubmatch(texto, -1)
	if len(itens) == 0 {
		return nil
	}

	produtos := make([]models.ProdXML, 0, len(itens))
	for _, item := range itens {
		nome := strings.TrimSpace(item[2])
		if nome == "" {
			continue
		}

		produtos = append(produtos, models.ProdXML{
			Nome:       nome,
			Quantidade: paraNumeroDecimal(strings.TrimSpace(item[3])),
			Unidade:    normalizarUnidade(strings.TrimSpace(item[4])),
			ValorUnit:  paraNumeroDecimal(strings.TrimSpace(item[5])),
		})
	}

	return produtos
}

// interpretarFormatoCupomComIndicadorImposto casa o terceiro formato (ver
// padraoItemComIndicadorImposto). A unidade nesse formato vem sempre colada
// à quantidade (ex: "1un"), então diferente dos outros dois formatos, aqui
// ela é sempre "UN" fixo — não tem uma unidade diferente pra capturar.
func interpretarFormatoCupomComIndicadorImposto(texto string) []models.ProdXML {
	itens := padraoItemComIndicadorImposto.FindAllStringSubmatch(texto, -1)
	if len(itens) == 0 {
		return nil
	}

	produtos := make([]models.ProdXML, 0, len(itens))
	for _, item := range itens {
		nome := strings.TrimSpace(item[2])
		if nome == "" {
			continue
		}

		quantidade := 1.0
		if valor, err := strconv.ParseFloat(strings.TrimSpace(item[3]), 64); err == nil {
			quantidade = valor
		}
		// Se não deu pra interpretar o token como número (ex: o OCR leu
		// "l" ou "j" no lugar do dígito), fica com o padrão 1 acima — a
		// tela de revisão corrige se estiver errado.

		produtos = append(produtos, models.ProdXML{
			Nome:       nome,
			Quantidade: fmt.Sprintf("%g", quantidade),
			Unidade:    "UN",
			ValorUnit:  paraNumeroDecimal(strings.TrimSpace(item[4])),
		})
	}

	return produtos
}
=======
>>>>>>> 774d74c6f21a1477a584a0d74f8b14f40344279e
