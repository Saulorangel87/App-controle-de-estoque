package services

import "testing"

func TestValidarURLExigeHTTPSMesmoComDominioPermitido(t *testing.T) {
	dominiosOriginais := dominiosPermitidosNFCe
	dominiosPermitidosNFCe = map[string]bool{"sefaz.exemplo.gov.br": true}
	t.Cleanup(func() { dominiosPermitidosNFCe = dominiosOriginais })

	if _, err := validarURL("http://sefaz.exemplo.gov.br/consulta?p=1"); err != ErrURLInvalida {
		t.Fatalf("erro = %v; esperado ErrURLInvalida", err)
	}
	if _, err := validarURL("https://sefaz.exemplo.gov.br:8443/consulta?p=1"); err != ErrURLInvalida {
		t.Fatalf("erro = %v; esperado ErrURLInvalida", err)
	}
}

func TestValidarURLAceitaSomenteAllowlistHTTPS(t *testing.T) {
	dominiosOriginais := dominiosPermitidosNFCe
	dominiosPermitidosNFCe = map[string]bool{"sefaz.exemplo.gov.br": true}
	t.Cleanup(func() { dominiosPermitidosNFCe = dominiosOriginais })

	if _, err := validarURL("https://outro.exemplo/consulta"); err != ErrDominioNaoSuportado {
		t.Fatalf("erro = %v; esperado ErrDominioNaoSuportado", err)
	}
	if _, err := validarURL("https://sefaz.exemplo.gov.br/consulta?p=1"); err != nil {
		t.Fatalf("URL HTTPS válida foi rejeitada: %v", err)
	}
}
