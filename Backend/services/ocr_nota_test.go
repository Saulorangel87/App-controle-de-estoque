package services

import (
	"bytes"
	"image"
	"image/jpeg"
	"testing"
)

func imagemJPEGTeste(t *testing.T, largura, altura int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, largura, altura)), nil); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestValidarImagemRejeitaDimensaoExcessiva(t *testing.T) {
	if err := ValidarImagem(imagemJPEGTeste(t, maximoLadoImagemOCR+1, 1)); err == nil {
		t.Fatal("imagem acima do lado máximo foi aceita")
	}
}

func TestValidarImagemAceitaJPEGDentroDoLimite(t *testing.T) {
	if err := ValidarImagem(imagemJPEGTeste(t, 120, 80)); err != nil {
		t.Fatalf("imagem JPEG válida foi rejeitada: %v", err)
	}
}
