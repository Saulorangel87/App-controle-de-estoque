package services

import (
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

<<<<<<< HEAD
// psmPrints é o Page Segmentation Mode usado na primeira tentativa —
// "bloco uniforme de texto", bom pra prints/screenshots digitais (a
// maioria dos casos, e mais rápido por ser uma tentativa só por rotação).
const psmPrints = "6"

// psmsFotoPapel são tentados só se a primeira camada não achar nada —
// "4" (coluna única de tamanhos variáveis) e "11" (texto esparso) tendem a
// funcionar melhor que o "6" em foto de cupom físico, que tem colunas,
// bordas e inclinação que um "bloco uniforme" não modela bem.
var psmsFotoPapel = []string{"4", "11"}

=======
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
// ExtrairProdutosDeImagem recebe os bytes de uma foto ou print da nota
// fiscal, roda OCR (Tesseract, em português) pra extrair o texto, e
// interpreta esse texto procurando um dos formatos conhecidos (tela da
// SEFAZ ou cupom de papel impresso — ver interpretarTextoOCR).
//
// O valor TOTAL de cada item não é lido da imagem — é CALCULADO
// (quantidade × valor unitário) pelo restante do fluxo, que já faz essa
// conta. Isso evita depender do OCR acertar mais um número por item, já
// que o valor total geralmente aparece separado visualmente (alinhado à
// direita), o que é mais fácil de o OCR embaralhar com outro item.
//
// FUNCIONA MELHOR com um print da tela de confirmação da SEFAZ (texto
// digital nítido) do que com foto do cupom de papel térmico (mais difícil
<<<<<<< HEAD
// pra qualquer OCR, por causa do desbotamento/reflexo/amassado do papel) —
// por isso a segunda camada abaixo existe especificamente pra foto física.
=======
// pra qualquer OCR, por causa do desbotamento/reflexo/amassado do papel).
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
func ExtrairProdutosDeImagem(imagem []byte) ([]models.ProdXML, error) {
	imagemDecodificada, _, err := image.Decode(bytes.NewReader(imagem))
	if err != nil {
		// Formato de imagem não suportado ou arquivo corrompido — não tem
		// como rotacionar/pré-processar nesse caso, tenta OCR direto nos
		// bytes originais como última tentativa.
<<<<<<< HEAD
		texto, errOCR := rodarOCR(imagem, psmPrints)
		if errOCR != nil {
			return nil, errOCR
		}
		return finalizarExtracao(texto, interpretarTextoOCR(texto))
	}

	// Camada 1: PSM 6 nas 4 rotações — rápido (4 chamadas ao Tesseract),
	// cobre a maioria dos casos (prints/screenshots digitais).
	texto, produtos, rodou, ultimoErro := tentarComPSMs(imagemDecodificada, []string{psmPrints})

	// Camada 2: só entra em ação se a primeira não achou NENHUM item — foto
	// de papel físico tende a precisar de um PSM diferente. Mais lento (mais
	// 8 chamadas ao Tesseract: 4 rotações × 2 PSMs), então só roda quando
	// realmente precisa.
	if len(produtos) == 0 {
		texto2, produtos2, rodou2, ultimoErro2 := tentarComPSMs(imagemDecodificada, psmsFotoPapel)
		if len(produtos2) > len(produtos) {
			texto, produtos = texto2, produtos2
		} else if texto == "" && texto2 != "" {
			texto = texto2 // garante que sobra algum texto pro debug
		}
		rodou = rodou || rodou2
		if ultimoErro2 != nil {
			ultimoErro = ultimoErro2
		}
	}

	if !rodou {
		// Nenhuma combinação de rotação/PSM sequer conseguiu rodar o OCR
		// (todas deram erro) — isso é diferente de "rodou mas não achou os
		// itens". Devolve o erro real do Tesseract em vez de mascarar como
		// ErrOCRSemItens, senão fica impossível saber o que deu errado.
		if ultimoErro != nil {
			return nil, ultimoErro
		}
		return nil, ErrOCRFalhou
	}

	return finalizarExtracao(texto, produtos)
}

// tentarComPSMs roda o OCR nas 4 rotações possíveis, testando cada um dos
// PSMs informados, e devolve o melhor resultado (mais itens reconhecidos)
// encontrado. rodouAlgumaVez indica se pelo menos uma combinação chegou a
// rodar o Tesseract com sucesso (mesmo sem achar itens) — usado por quem
// chama pra decidir se o erro é "não rodou o OCR" ou "rodou mas não achou".
func tentarComPSMs(imagemDecodificada image.Image, psms []string) (
	melhorTexto string, melhoresProdutos []models.ProdXML, rodouAlgumaVez bool, ultimoErro error,
) {
=======
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

>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
	for rotacoes := 0; rotacoes < 4; rotacoes++ {
		imagemRotacionada := rotacionar90(imagemDecodificada, rotacoes)

		imagemProcessada, err := prepararImagemParaOCR(imagemRotacionada)
		if err != nil {
			ultimoErro = err
			continue
		}

<<<<<<< HEAD
		for _, psm := range psms {
			texto, err := rodarOCR(imagemProcessada, psm)
			if err != nil {
				ultimoErro = err
				continue
			}

			if !rodouAlgumaVez {
				melhorTexto = texto
				rodouAlgumaVez = true
			}

			produtos := interpretarTextoOCR(texto)
			if len(produtos) > len(melhoresProdutos) {
				melhoresProdutos = produtos
				melhorTexto = texto
			}
		}
	}
	return
=======
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
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
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
		return nil, ErrOCRSemItens
	}

	return produtos, nil
}

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
<<<<<<< HEAD
func rodarOCR(imagemProcessada []byte, psm string) (string, error) {
=======
func rodarOCR(imagemProcessada []byte) (string, error) {
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
	arquivoTemp, err := os.CreateTemp("", "nota-ocr-*.png")
	if err != nil {
		return "", ErrOCRFalhou
	}
	defer os.Remove(arquivoTemp.Name())

	if _, err := arquivoTemp.Write(imagemProcessada); err != nil {
		arquivoTemp.Close()
		return "", ErrOCRFalhou
	}
	arquivoTemp.Close()

	// "-l por": usa o idioma português (precisa do pacote de idioma
	// instalado junto com o tesseract-ocr, ver Backend.Dockerfile).
<<<<<<< HEAD
	// "--psm": ver psmPrints/psmsFotoPapel — modo de segmentação de página,
	// varia conforme a origem provável da imagem (tela digital vs papel).
=======
	// "--psm 6": trata a imagem como um bloco uniforme de texto — funciona
	// melhor pra telas/listas como essa do que o modo automático padrão,
	// que é pensado pra páginas de documento com parágrafos "de verdade".
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
	// "--dpi 300": screenshots normalmente não têm informação de DPI
	// embutida, e sem isso o Tesseract às vezes assume uma resolução baixa
	// e piora a leitura de números pequenos — forçar 300 evita essa
	// suposição errada.
	// "stdout": manda o texto reconhecido direto pra saída padrão, sem
	// precisar gerenciar mais um arquivo temporário de resultado.
	saida, err := exec.Command(
		"tesseract", arquivoTemp.Name(), "stdout",
<<<<<<< HEAD
		"-l", "por", "--psm", psm, "--dpi", "300",
=======
		"-l", "por", "--psm", "6", "--dpi", "300",
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
	).Output()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOCRFalhou, err)
	}

	return string(saida), nil
}

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
<<<<<<< HEAD
// caso, já rotacionada — ver rotacionar90), converte pra escala de cinza,
// amplia a resolução e binariza (preto e branco puro, sem cinza) usando o
// método de Otsu, devolvendo um novo PNG.
//
// Escala de cinza reduz ruído de cor que não ajuda em nada o
// reconhecimento de texto; ampliar dá mais definição pros traços finos dos
// caracteres; binarizar remove o "cinza" residual (sombra, papel
// amarelado, compressão JPEG) que confunde o Tesseract — ele funciona bem
// melhor com preto/branco puro do que com tons de cinza intermediários.
=======
// caso, já rotacionada — ver rotacionar90), converte pra escala de cinza e
// amplia a resolução, devolvendo um novo PNG. Escala de cinza reduz ruído
// de cor que não ajuda em nada o reconhecimento de texto; ampliar dá mais
// definição pros traços finos dos caracteres.
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
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

<<<<<<< HEAD
	imagemBinarizada := binarizarOtsu(imagemAmpliada)

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imagemBinarizada); err != nil {
=======
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imagemAmpliada); err != nil {
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
		return nil, err
	}

	return buffer.Bytes(), nil
}

<<<<<<< HEAD
// binarizarOtsu converte uma imagem em escala de cinza pra preto e branco
// puro, escolhendo automaticamente o limiar que melhor separa "texto" de
// "fundo" (método de Otsu: maximiza a variância entre as duas classes de
// pixels). É um limiar GLOBAL (um valor só pra imagem inteira) — funciona
// bem na maioria dos casos, mas uma foto com iluminação muito desigual
// (sombra forte de um lado, por exemplo) se beneficiaria de um limiar
// adaptativo por região, que é mais complexo e não foi necessário até agora.
func binarizarOtsu(img *image.Gray) *image.Gray {
	limites := img.Bounds()

	var histograma [256]int
	for y := limites.Min.Y; y < limites.Max.Y; y++ {
		for x := limites.Min.X; x < limites.Max.X; x++ {
			histograma[img.GrayAt(x, y).Y]++
		}
	}

	total := limites.Dx() * limites.Dy()
	var somaTotal float64
	for nivel, contagem := range histograma {
		somaTotal += float64(nivel * contagem)
	}

	var somaFundo float64
	var pesoFundo int
	var melhorVariancia float64
	melhorLimiar := 0

	for limiar := 0; limiar < 256; limiar++ {
		pesoFundo += histograma[limiar]
		if pesoFundo == 0 {
			continue
		}
		pesoFrente := total - pesoFundo
		if pesoFrente == 0 {
			break
		}

		somaFundo += float64(limiar * histograma[limiar])
		mediaFundo := somaFundo / float64(pesoFundo)
		mediaFrente := (somaTotal - somaFundo) / float64(pesoFrente)
		diferenca := mediaFundo - mediaFrente

		variancia := float64(pesoFundo) * float64(pesoFrente) * diferenca * diferenca
		if variancia > melhorVariancia {
			melhorVariancia = variancia
			melhorLimiar = limiar
		}
	}

	resultado := image.NewGray(limites)
	for y := limites.Min.Y; y < limites.Max.Y; y++ {
		for x := limites.Min.X; x < limites.Max.X; x++ {
			if img.GrayAt(x, y).Y >= uint8(melhorLimiar) {
				resultado.SetGray(x, y, color.Gray{Y: 255})
			} else {
				resultado.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	return resultado
}

=======
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
// padraoNomeCodigo casa linhas como "LEITE INTEGRAL GODAM (Código: 22080 )".
// (?m) trata cada linha do texto separadamente. Aceita pequenas variações
// que o OCR costuma introduzir na palavra "Código" (0 no lugar do O, etc.)
// e na pontuação ao redor do número.
var padraoNomeCodigo = regexp.MustCompile(`(?m)^(.+?)\s*\(C[oó0]digo:?\s*(\d+)\s*\)`)

// padraoQuantidade casa trechos como "Qtde.:1 UN: UN Vl. Unit.: 5,89" — os
// rótulos ("Qtde", "UN", "Vl. Unit") são fixos no template da SEFAZ, então
// são um ponto de ancoragem confiável mesmo se a pontuação ao redor variar
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

<<<<<<< HEAD
// padraoLinhaNaoProduto reconhece linhas de rodapé/totalização que às vezes
// têm números parecidos o bastante com preço/quantidade pra confundir os
// parsers de item — remover essas linhas antes de tentar interpretar deixa
// o texto mais "limpo" pra qualquer um dos formatos.
var padraoLinhaNaoProduto = regexp.MustCompile(
	`(?mi)^.*\b(TOTAL|SUBTOTAL|TROCO|DINHEIRO|CART[ÃA]O|PIX|DESCONTO|ACR[ÉE]SCIMO|TRIBUTOS?|CNPJ|CPF|DATA|HORA|SAT|NFC-?e|OBRIGADO|VALOR A PAGAR|FORMA DE PAGAMENTO|PROTOCOLO|CHAVE DE ACESSO|CONSULTE)\b.*$`,
)

// removerLinhasNaoProduto apaga (deixa em branco) as linhas reconhecidas
// como rodapé/totalização, mantendo as outras intactas.
func removerLinhasNaoProduto(texto string) string {
	return padraoLinhaNaoProduto.ReplaceAllString(texto, "")
}

=======
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
// interpretarTextoOCR tenta reconhecer, em ordem, os formatos conhecidos de
// nota/cupom — cada rede/portal usa um layout diferente, então em vez de
// uma tentativa só, tenta cada um até algum encontrar itens:
//  1. Tela da SEFAZ (mais comum quando a origem é um print/screenshot)
//  2. Cupom de papel com "qtd unidade x preço total"
//  3. Cupom de papel com número do item + indicador de imposto
func interpretarTextoOCR(texto string) []models.ProdXML {
<<<<<<< HEAD
	texto = removerLinhasNaoProduto(texto)

=======
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
	if produtos := interpretarFormatoTelaSefaz(texto); len(produtos) > 0 {
		return produtos
	}
	if produtos := interpretarFormatoCupomPapel(texto); len(produtos) > 0 {
		return produtos
	}
	return interpretarFormatoCupomComIndicadorImposto(texto)
}

<<<<<<< HEAD
// interpretarFormatoTelaSefaz casa cada "nome + código" com a "quantidade +
// unidade + valor" que aparece DEPOIS dele e ANTES do próximo nome (uma
// janela de posição no texto) — o template da tela da SEFAZ sempre
// intercala essas duas linhas por item.
//
// Por que por JANELA e não pareando os arrays por índice: se o OCR "engole"
// a linha de quantidade de um item no meio da lista, parear por índice
// desalinha TODOS os itens seguintes silenciosamente (o item 5 herdaria a
// quantidade do item 4, e assim por diante) — um bug muito pior do que
// simplesmente faltar a quantidade de UM item. Pareando por janela, um item
// sem quantidade encontrada fica só com o padrão (1 UN, sem valor) — os
// outros continuam corretos, e a tela de revisão sinaliza esse caso
// específico pra correção manual.
func interpretarFormatoTelaSefaz(texto string) []models.ProdXML {
	nomesIdx := padraoNomeCodigo.FindAllStringSubmatchIndex(texto, -1)
	if len(nomesIdx) == 0 {
		return nil
	}
	quantidadesIdx := padraoQuantidade.FindAllStringSubmatchIndex(texto, -1)

	produtos := make([]models.ProdXML, 0, len(nomesIdx))
	cursorQuantidade := 0

	for i, nomeMatch := range nomesIdx {
		nome := strings.TrimSpace(texto[nomeMatch[2]:nomeMatch[3]])
=======
// interpretarFormatoTelaSefaz casa cada "nome + código" (na ordem em que
// aparecem no texto) com a "quantidade + unidade + valor" correspondente
// (também na ordem em que aparecem) — o template da tela da SEFAZ sempre
// intercala essas duas linhas por item, então parear PELA ORDEM de
// aparição é mais confiável do que tentar casar tudo numa regex só, já que
// o OCR quebra linha de forma imprevisível dependendo da imagem.
func interpretarFormatoTelaSefaz(texto string) []models.ProdXML {
	nomes := padraoNomeCodigo.FindAllStringSubmatch(texto, -1)
	quantidades := padraoQuantidade.FindAllStringSubmatch(texto, -1)

	// Se as contagens não baterem, o OCR deve ter engolido ou duplicado
	// alguma linha — melhor devolver nada (interpretarTextoOCR cai pro
	// formato de cupom de papel, e se esse também não achar nada, o
	// handler cai no erro ErrOCRSemItens) do que arriscar parear um item
	// com a quantidade errada silenciosamente.
	if len(nomes) == 0 || len(nomes) != len(quantidades) {
		return nil
	}

	produtos := make([]models.ProdXML, 0, len(nomes))
	for i := range nomes {
		nome := strings.TrimSpace(nomes[i][1])
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
		if nome == "" {
			continue
		}

<<<<<<< HEAD
		inicioJanela := nomeMatch[1] // logo após o fim do match do nome
		fimJanela := len(texto)
		if i+1 < len(nomesIdx) {
			fimJanela = nomesIdx[i+1][0] // início do próximo nome
		}

		// Avança o cursor até a primeira quantidade que esteja dentro (ou
		// depois d)o início da janela deste item — quantidades de itens
		// anteriores já foram consumidas e nunca são reutilizadas.
		for cursorQuantidade < len(quantidadesIdx) && quantidadesIdx[cursorQuantidade][0] < inicioJanela {
			cursorQuantidade++
		}

		item := models.ProdXML{Nome: nome, Quantidade: "1", Unidade: "UN"}

		if cursorQuantidade < len(quantidadesIdx) && quantidadesIdx[cursorQuantidade][0] < fimJanela {
			q := quantidadesIdx[cursorQuantidade]

			quantidadeLida := strings.ReplaceAll(strings.TrimSpace(texto[q[2]:q[3]]), "?", "")
			if quantidadeLida == "" {
				// O dígito sumiu do OCR — assume 1 (o mais comum em
				// compras), a revisão manual corrige se estiver errado.
				quantidadeLida = "1"
			}

			item.Quantidade = paraNumeroDecimal(quantidadeLida)
			item.Unidade = normalizarUnidade(strings.TrimSpace(texto[q[4]:q[5]]))
			item.ValorUnit = paraNumeroDecimal(strings.TrimSpace(texto[q[6]:q[7]]))

			cursorQuantidade++ // essa quantidade já foi usada
		}
		// Se não achou nenhuma quantidade dentro da janela deste item, o
		// item entra com o padrão (1 UN, sem valor) em vez de ser
		// descartado — melhor pedir revisão de um item do que perder a
		// nota inteira por causa de uma linha que o OCR engoliu.

		produtos = append(produtos, item)
=======
		quantidadeLida := strings.ReplaceAll(strings.TrimSpace(quantidades[i][1]), "?", "")
		if quantidadeLida == "" {
			// O dígito sumiu do OCR — assume 1 (o mais comum em compras),
			// e conta com a revisão manual pra corrigir se estiver errado.
			quantidadeLida = "1"
		}

		produtos = append(produtos, models.ProdXML{
			Nome:       nome,
			Quantidade: paraNumeroDecimal(quantidadeLida),
			Unidade:    normalizarUnidade(strings.TrimSpace(quantidades[i][2])),
			ValorUnit:  paraNumeroDecimal(strings.TrimSpace(quantidades[i][3])),
		})
>>>>>>> 482a404a93b8b431c81c28e1caa87cd9ec93d222
	}

	return produtos
}

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
