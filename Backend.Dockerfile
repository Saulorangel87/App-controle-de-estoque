# Estágio 1: compila o binário Go num ambiente completo com as ferramentas de build
FROM golang:1.26.5-alpine AS builder
WORKDIR /app

# Copia primeiro só os arquivos de dependência — o Docker cacheia essa camada,
# então "go mod download" só roda de novo se go.mod/go.sum mudarem, não a cada build.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 gera um binário estático (sem depender de libs do sistema),
# essencial porque a imagem final (alpine) não tem as mesmas libs do builder.
RUN CGO_ENABLED=0 GOOS=linux go build -o servidor .

# Estágio 2: imagem final, só com o binário compilado — bem menor que carregar
# toda a toolchain do Go em produção.
FROM alpine:3.20
WORKDIR /app

# ca-certificates é necessário para chamadas HTTPS de saída (não usamos hoje,
# mas evita surpresa se alguma dependência futura precisar).
# tesseract-ocr + tesseract-ocr-data-por: usados pela importação de nota
# fiscal via foto/print (OCR) — o binário "tesseract" é chamado via
# exec.Command pelo backend, não é uma lib Go, então precisa estar instalado
# na imagem final.
RUN apk add --no-cache ca-certificates tesseract-ocr tesseract-ocr-data-por

COPY --from=builder /app/servidor .

EXPOSE 8080
CMD ["./servidor"]
