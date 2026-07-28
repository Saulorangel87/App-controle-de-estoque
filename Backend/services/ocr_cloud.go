package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"controle-estoque/models"
)

const (
	ocrSpaceURL = "https://api.ocr.space/parse/image"

	// ocrSpaceLimiteBytes fica abaixo do limite real de 1MB do plano
	// grátis do OCR.space — margem de segurança pra não estourar por causa
	// do overhead do multipart/form-data em volta dos bytes da imagem.
	ocrSpaceLimiteBytes = 900 * 1024

	// ladoMaximoOCRSpace: fotos de celular modernas costumam ter resolução
	// muito maior do que qualquer OCR precisa — reduzir antes de comprimir
	// mantém a legibilidade do texto e ainda ajuda a caber no limite de
	// tamanho.
	ladoMaximoOCRSpace = 2000
)

var (
	ErrOCRCloudSemChave = errors.New(
		"a leitura de foto de cupom de papel não está configurada no momento — envie o XML da nota ou um print da tela da SEFAZ",
	)
	ErrOCRCloudFalhou = errors.New(
		"não foi possível consultar o serviço de leitura de imagem agora — tente novamente em instantes",
	)
)

// ExtrairProdutosDeImagemViaOCRSpace é o equivalente de ExtrairProdutosDeImagem
// (ver ocr_nota.go), mas usa a API de nuvem do OCR.space em vez do Tesseract
// local. Usado especificamente pra FOTO DE CUPOM DE PAPEL FÍSICO — prints da
// tela da SEFAZ continuam passando pelo Tesseract (rápido, de graça, e já
// funciona bem pra esse caso).
//
// Por que só pra foto de papel: fotos reais de cupom (ângulo, iluminação,
// papel térmico desbotado) se beneficiam MUITO de um motor de OCR mais
// robusto — a API já corrige rotação e ruído de fundo sozinha, o que o
// Tesseract local não faz tão bem mesmo com pré-processamento.
//
// O texto retornado pela API passa pelos MESMOS parsers (interpretarTextoOCR)
// usados pro Tesseract — a diferença é só a qualidade do texto de entrada,
// não a lógica de extrair produtos dele.
func ExtrairProdutosDeImagemViaOCRSpace(imagem []byte, apiKey string) ([]models.ProdXML, error) {
	if apiKey == "" {
		return nil, ErrOCRCloudSemChave
	}

	imagemPreparada, err := comprimirParaLimiteOCRSpace(imagem)
	if err != nil {
		// Falha ao comprimir (formato inesperado, etc.) — tenta com a
		// imagem original mesmo assim; se for grande demais, a própria
		// API vai recusar com uma mensagem clara.
		imagemPreparada = imagem
	}

	texto, err := chamarOCRSpace(imagemPreparada, apiKey)
	if err != nil {
		return nil, err
	}

	produtos := interpretarTextoOCR(texto)
	if len(produtos) == 0 {
		// Mesmo debug usado pelo fluxo do Tesseract — mesma mecânica,
		// mesma forma de diagnosticar se o texto veio limpo mas o parser
		// não reconheceu o formato.
		salvarDebugOCR(texto)
		return nil, ErrOCRSemItens
	}

	return produtos, nil
}

// chamarOCRSpace monta a requisição multipart pra API e devolve o texto
// reconhecido. Parâmetros escolhidos especificamente pro caso de foto de
// cupom de papel (ver documentação da API):
//   - OCREngine=2: bom equilíbrio geral, reconhece bem texto em fundo
//     "ruidoso" (foto) e lida melhor com texto rotacionado.
//   - detectOrientation=true: a própria API endireita a imagem sozinha —
//     não precisamos do nosso loop de 4 rotações nesse fluxo.
//   - scale=true: upscaling interno da API, ajuda em texto pequeno.
//   - isTable=true: recomendado pela documentação especificamente pra OCR
//     de cupom/recibo — garante que o texto volte organizado linha a linha.
func chamarOCRSpace(imagem []byte, apiKey string) (string, error) {
	var corpo bytes.Buffer
	escritor := multipart.NewWriter(&corpo)

	campoArquivo, err := escritor.CreateFormFile("file", "nota.jpg")
	if err != nil {
		return "", ErrOCRCloudFalhou
	}
	if _, err := campoArquivo.Write(imagem); err != nil {
		return "", ErrOCRCloudFalhou
	}

	_ = escritor.WriteField("language", "por")
	_ = escritor.WriteField("OCREngine", "2")
	_ = escritor.WriteField("detectOrientation", "true")
	_ = escritor.WriteField("scale", "true")
	_ = escritor.WriteField("isTable", "true")

	if err := escritor.Close(); err != nil {
		return "", ErrOCRCloudFalhou
	}

	requisicao, err := http.NewRequest(http.MethodPost, ocrSpaceURL, &corpo)
	if err != nil {
		return "", ErrOCRCloudFalhou
	}
	requisicao.Header.Set("apikey", apiKey)
	requisicao.Header.Set("Content-Type", escritor.FormDataContentType())

	// Timeout generoso — OCR na nuvem de uma imagem pode levar alguns
	// segundos, principalmente em fotos grandes.
	// Timeout mais curto que o normal de propósito: melhor o NOSSO código
	// desistir com uma mensagem clara em ~25s do que deixar o Cloudflare
	// Tunnel (ou outro proxy no meio do caminho) derrubar a conexão sem
	// aviso depois de esperar mais tempo — isso aparece no navegador como
	// "Failed to fetch", sem nenhuma mensagem útil pro usuário.
	cliente := &http.Client{Timeout: 25 * time.Second}
	resposta, err := cliente.Do(requisicao)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOCRCloudFalhou, err)
	}
	defer resposta.Body.Close()

	corpoResposta, err := io.ReadAll(resposta.Body)
	if err != nil {
		return "", ErrOCRCloudFalhou
	}

	var resultado struct {
		ParsedResults []struct {
			ParsedText        string `json:"ParsedText"`
			FileParseExitCode int    `json:"FileParseExitCode"`
		} `json:"ParsedResults"`
		IsErroredOnProcessing bool `json:"IsErroredOnProcessing"`
		// ErrorMessage no nível raiz às vezes vem como string, às vezes
		// como lista de strings (comportamento documentado de forma
		// inconsistente pela própria API) — usar interface{} evita que o
		// unmarshal quebre por causa disso; não precisamos exibir o texto
		// exato pro usuário mesmo, só decidir se deu erro.
		ErrorMessage interface{} `json:"ErrorMessage"`
	}

	if err := json.Unmarshal(corpoResposta, &resultado); err != nil {
		return "", fmt.Errorf("%w (resposta inesperada da API)", ErrOCRCloudFalhou)
	}

	if resultado.IsErroredOnProcessing || len(resultado.ParsedResults) == 0 {
		return "", ErrOCRCloudFalhou
	}

	return resultado.ParsedResults[0].ParsedText, nil
}

// comprimirParaLimiteOCRSpace reduz resolução e/ou qualidade JPEG até a
// imagem caber no limite de tamanho do plano grátis do OCR.space. Se a
// imagem já está dentro do limite, devolve ela sem recodificar — evita
// perda de qualidade desnecessária.
func comprimirParaLimiteOCRSpace(imagemOriginal []byte) ([]byte, error) {
	if len(imagemOriginal) <= ocrSpaceLimiteBytes {
		return imagemOriginal, nil
	}

	imagemDecodificada, _, err := image.Decode(bytes.NewReader(imagemOriginal))
	if err != nil {
		return nil, err
	}

	imagemRedimensionada := limitarResolucao(imagemDecodificada, ladoMaximoOCRSpace)

	// Tenta qualidades decrescentes até caber no limite — prioriza manter
	// mais qualidade possível, só reduz mais se precisar.
	for qualidade := 85; qualidade >= 40; qualidade -= 15 {
		var buffer bytes.Buffer
		if err := jpeg.Encode(&buffer, imagemRedimensionada, &jpeg.Options{Quality: qualidade}); err != nil {
			return nil, err
		}
		if buffer.Len() <= ocrSpaceLimiteBytes {
			return buffer.Bytes(), nil
		}
	}

	// Nem no menor nível de qualidade testado coube no limite — devolve o
	// melhor esforço mesmo assim (se a API recusar por tamanho, o erro
	// dela chega como ErrOCRCloudFalhou, com mensagem clara pro usuário).
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, imagemRedimensionada, &jpeg.Options{Quality: 40}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// limitarResolucao reduz a imagem (amostragem vizinho mais próximo, mesma
// técnica usada na ampliação em ocr_nota.go) se o lado maior ultrapassar
// ladoMaximo — mantém a proporção original.
func limitarResolucao(img image.Image, ladoMaximo int) image.Image {
	limites := img.Bounds()
	largura, altura := limites.Dx(), limites.Dy()

	maiorLado := largura
	if altura > maiorLado {
		maiorLado = altura
	}
	if maiorLado <= ladoMaximo {
		return img
	}

	fator := float64(ladoMaximo) / float64(maiorLado)
	novaLargura := int(float64(largura) * fator)
	novaAltura := int(float64(altura) * fator)

	redimensionada := image.NewRGBA(image.Rect(0, 0, novaLargura, novaAltura))
	for y := 0; y < novaAltura; y++ {
		yOriginal := limites.Min.Y + int(float64(y)/fator)
		for x := 0; x < novaLargura; x++ {
			xOriginal := limites.Min.X + int(float64(x)/fator)
			redimensionada.Set(x, y, img.At(xOriginal, yOriginal))
		}
	}
	return redimensionada
}
