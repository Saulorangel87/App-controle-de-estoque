package config

import (
	"errors"
	"os"
	"strings"
)

const tamanhoMinimoChaveJWT = 32

// Validar verifica configurações obrigatórias antes de o servidor aceitar
// requisições autenticadas. Um segredo vazio ou curto tornaria possível forjar
// tokens com facilidade, então o backend deve falhar ao iniciar nesse caso.
func Validar() error {
	segredo := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if segredo == "" {
		return errors.New("JWT_SECRET não configurado")
	}
	if len([]byte(segredo)) < tamanhoMinimoChaveJWT {
		return errors.New("JWT_SECRET deve ter pelo menos 32 bytes")
	}
	return nil
}

// ChaveSecreta e ChaveOCRSpace são FUNÇÕES (não variáveis) de propósito:
// variável de pacote é avaliada na inicialização do programa, ANTES da
// func main() começar a rodar — e é lá que o main.go chama
// godotenv.Load() pra carregar o .env. Se fossem variáveis, elas sempre
// leriam os.Getenv(...) vazio (o .env ainda não tinha sido carregado nesse
// momento), mesmo com o valor certo no arquivo. Como função, a leitura só
// acontece quando alguém chama ChaveSecreta()/ChaveOCRSpace() de verdade —
// nesse ponto, main() já rodou o godotenv.Load() e os.Getenv já enxerga o
// valor certo.
func ChaveSecreta() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// ChaveOCRSpace é usada só no fluxo de importação por FOTO DE PAPEL físico
// (o print da tela da SEFAZ continua usando o Tesseract local, de graça).
// Se estiver vazia, esse fluxo específico fica indisponível com uma
// mensagem clara — ver services/ocr_cloud.go.
func ChaveOCRSpace() string {
	return os.Getenv("OCR_SPACE_API_KEY")
}

// DebugOCRAtivo é deliberadamente falso por padrão. Os textos reconhecidos e
// HTML de notas podem conter dados fiscais ou pessoais e só devem ser gravados
// durante uma investigação local explicitamente autorizada.
func DebugOCRAtivo() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("DEBUG_OCR")), "true")
}
