# Estágio 1: instala dependências e gera o build estático de produção
FROM node:20-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

# VITE_API_URL precisa existir NO MOMENTO DO BUILD, não em tempo de execução —
# o Vite substitui import.meta.env.VITE_API_URL pelo valor literal dentro do
# JavaScript já compilado. Por isso vem como build arg, não como variável comum
# do container (que só existiria depois que o build já terminou).
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# Estágio 2: serve os arquivos estáticos gerados com nginx — muito mais leve
# do que manter o Node rodando em produção só para servir arquivos prontos.
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
