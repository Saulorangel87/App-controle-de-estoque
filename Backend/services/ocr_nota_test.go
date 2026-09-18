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

func TestFatorSeguroAmpliacaoReduzImagensGrandes(t *testing.T) {
	if fator := fatorSeguroAmpliacao(4000, 3000); fator != 1 {
		t.Fatalf("fator para imagem de 12 MP = %d; esperado 1", fator)
	}
	if fator := fatorSeguroAmpliacao(2000, 2000); fator != 3 {
		t.Fatalf("fator para imagem de 4 MP = %d; esperado 3", fator)
	}
}
