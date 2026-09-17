package config

import "testing"

func TestValidarJWTSecret(t *testing.T) {
	tests := []struct {
		nome       string
		segredo    string
		deveFalhar bool
	}{
		{nome: "ausente", segredo: "", deveFalhar: true},
		{nome: "curto", segredo: "1234567890123456789012345678901", deveFalhar: true},
		{nome: "válido", segredo: "12345678901234567890123456789012", deveFalhar: false},
	}

	for _, teste := range tests {
		t.Run(teste.nome, func(t *testing.T) {
			t.Setenv("JWT_SECRET", teste.segredo)
			err := Validar()
			if (err != nil) != teste.deveFalhar {
				t.Fatalf("Validar() = %v; deveFalhar = %v", err, teste.deveFalhar)
			}
		})
	}
}
