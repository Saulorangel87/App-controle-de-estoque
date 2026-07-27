package config

import "os"

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
