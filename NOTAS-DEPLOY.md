# 📦 Notas de Deploy — Controle de Estoque

---

# 🧱 Stack do Projeto

## Backend
- Go (`net/http` puro)
- SQLite
- `modernc.org/sqlite`
- bcrypt para criptografia de senhas
- JWT para autenticação

## Frontend
- React 19
- Vite
- React Router DOM
- CSS puro (sem bibliotecas de UI)

## Deploy
- Docker Compose
- Backend + Frontend com Nginx
- Oracle Cloud
- Cloudflare Tunnel

---

# 📁 Estrutura do Projeto

```
App-controle-de-estoque/

├── docker-compose.yml
├── Backend.Dockerfile
├── Frontend.Dockerfile
├── nginx.conf
├── .gitignore

├── Backend/
│   ├── main.go
│   ├── config/
│   │   └── chave JWT via variável de ambiente
│   │
│   ├── database/
│   │   └── conexão SQLite + migração de colunas
│   │
│   ├── models/
│   │
│   ├── handlers/
│   │   └── autenticação + itens
│   │
│   └── middleware/
│       └── autenticação JWT

└── Frontend/

    ├── src/
    │
    ├── api/
    │   └── cliente HTTP centralizado
    │
    ├── context/
    │   └── Auth + Theme
    │
    ├── components/
    │   ├── Footer
    │   ├── modais
    │   ├── tabela
    │   ├── filtros
    │   └── cards
    │
    ├── pages/
    │   ├── Login
    │   └── Dashboard
    │
    └── utils/
        ├── JWT decode
        └── perguntas de segurança
```

---

# ✅ Funcionalidades Implementadas

## 🔐 Autenticação

- Cadastro de usuário:
  - Usuário
  - Senha
  - Pergunta de segurança

- Login utilizando JWT

- Recuperação de senha através de pergunta de segurança em 2 etapas

---

## 📦 Controle de Estoque

- CRUD completo de itens:

  - Adicionar
  - Listar
  - Editar
  - Excluir
  - Retirar (baixa de estoque)

- Controle de quantidade

- Definição de estoque mínimo

- Alerta de estoque baixo:

```
Quantidade <= Estoque mínimo
```

---

## 📊 Dashboard

- Cards de resumo:

  - Total de itens
  - Produtos com estoque baixo
  - Localizações em uso

---

## 🔎 Busca e Filtros

- Busca por nome

- Filtro por localização:

  - Despensa
  - Geladeira
  - Freezer
  - Armário

---

## 🎨 Interface

- Tema claro/escuro
- Persistência do tema utilizando `localStorage`
- Nome da conta exibido no cabeçalho:
  - Extraído do JWT
  - Exibe os dois primeiros nomes

---

## 📝 Rodapé

- Fixo na base da tela

Informações:

- Copyright
- Versão (`v1.0.0`)
- Links:
  - LinkedIn
  - GitHub
  - Email

---

# ♿ Boas Práticas Aplicadas

- Código comentado
- Cuidados de acessibilidade:

  - Labels
  - Foco visível
  - `aria-live`
  - `aria-modal`

- Cuidados de performance:

  - Sem fontes externas
  - Sem bibliotecas pesadas de UI

---

# 🚀 Deploy — Em Andamento

## Configuração atual

- Dockerfiles preparados
- `docker-compose.yml` configurado

Serviços:

```
Backend
Porta: 8090

Frontend
Porta: 8091
```

---

## Variáveis configuráveis

### Backend

- CORS configurável
- Caminho do banco via:

```
DB_PATH
```

- Chave JWT via variável de ambiente

---

### Frontend

URL da API configurável através de:

```
VITE_API_URL
```

---

## Domínios planejados

Frontend:

```
estoque.devsaulo.com.br
```

Backend:

```
estoque-api.devsaulo.com.br
```

---

## Pendências de Deploy

- Gerar chave JWT de produção
- Subir projeto para GitHub
- Executar na VM:

```bash
git pull
```

Depois:

```bash
docker compose up -d --build
```

- Apontar subdomínios no:

  - Nginx Proxy Manager
  - Cloudflare

---

# ⚠️ Pendências Conhecidas

## Interface Mobile

Os ícones de:

- Editar
- Excluir

na tabela precisam de ajustes para telas menores.

Alteração necessária:

- Manter os ícones pequenos
- Deixar ao lado do botão "Retirar"
- Evitar corte ou desaparecimento

---

# 🔒 Segurança

## Variável JWT

A chave:

```
ChaveSecreta
```

deve ser alterada caso exista qualquer possibilidade de vazamento.

Regra:

- Nunca reutilizar a chave local em produção.
- Cada ambiente deve possuir sua própria chave.

---

# 📌 Arquivos ignorados no Git

## Backend

```
Backend/.env
Backend/estoque.db
```

## Frontend

```
Frontend/node_modules
Frontend/dist
Frontend/.env
```

## Produção

```
data/
```

Observação:

```
data/
```

existe somente na VM de produção e nunca deve ser enviado para o GitHub.

---

# 📅 Versão das Notas

```
Deploy v1.0.0
```
