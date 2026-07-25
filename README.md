# 📦 Controle de Estoque

Aplicação web completa para gerenciamento de estoque doméstico ou empresarial, permitindo cadastro de usuários, autenticação segura e controle de itens com alertas de estoque baixo.

Projeto desenvolvido com separação entre **Frontend** e **Backend**, utilizando uma API própria e banco SQLite.

## 🔗 Projeto no ar

Acesse a aplicação:

🔗 [Projeto no ar, acesse aqui](https://estoque.devsaulo.com.br)

---

## 🚀 Tecnologias utilizadas

### Backend
- Go
- `net/http` (API REST sem frameworks)
- SQLite
- `modernc.org/sqlite`
- JWT para autenticação
- bcrypt para criptografia de senhas

### Frontend
- React 19
- Vite
- React Router DOM
- CSS puro
- Context API
- LocalStorage para persistência de preferências

### Deploy
- Docker
- Docker Compose
- Nginx
- Oracle Cloud
- Cloudflare Tunnel

---

# 📁 Estrutura do projeto

```
App-controle-de-estoque/
│
├── docker-compose.yml
├── Backend.Dockerfile
├── Frontend.Dockerfile
├── nginx.conf
├── .gitignore
│
├── Backend/
│   ├── main.go
│   ├── config/
│   ├── database/
│   ├── models/
│   ├── handlers/
│   └── middleware/
│
└── Frontend/
    ├── src/
    │   ├── api/
    │   ├── components/
    │   ├── context/
    │   ├── pages/
    │   └── utils/
    │
    └── public/
```

---

# ✨ Funcionalidades

## 🔐 Autenticação
- Cadastro de usuários
- Login com JWT
- Senhas protegidas com bcrypt
- Recuperação de senha através de pergunta de segurança
- Controle de sessão utilizando token JWT

---

## 📦 Controle de estoque

- Cadastro de produtos
- Listagem de itens
- Edição de produtos
- Exclusão de produtos
- Retirada de estoque (baixa automática)
- Controle de quantidade
- Definição de estoque mínimo

---

## 📊 Dashboard

- Total de itens cadastrados
- Quantidade de produtos em estoque baixo
- Locais utilizados
- Indicadores rápidos do estoque

---

## 🔎 Busca e filtros

- Pesquisa por nome
- Filtro por localização:

  - Despensa
  - Geladeira
  - Freezer
  - Armário

---

## 🎨 Interface

- Tema claro e escuro
- Preferência salva no navegador
- Layout responsivo
- Componentes reutilizáveis
- Feedback visual para ações do usuário

---

# ♿ Boas práticas aplicadas

- Labels associados aos campos
- Foco visível para navegação
- Uso de `aria-live`
- Uso de `aria-modal`
- Código organizado por responsabilidade
- Sem dependência de bibliotecas pesadas de UI

---

# 🛠️ Executando localmente

## Pré-requisitos

Instale:

- Go
- Node.js
- Docker (opcional)

---

## Backend

Entre na pasta:

```bash
cd Backend
```

Configure as variáveis de ambiente criando um arquivo:

```
.env
```

Exemplo:

```env
JWT_SECRET=sua_chave_secreta
DB_PATH=estoque.db
```

Execute:

```bash
go run main.go
```

O backend ficará disponível na porta configurada.

---

## Frontend

Entre na pasta:

```bash
cd Frontend
```

Instale as dependências:

```bash
npm install
```

Configure:

```
.env
```

Exemplo:

```env
VITE_API_URL=http://localhost:8090
```

Execute:

```bash
npm run dev
```

---

# 🐳 Executando com Docker

O projeto possui configuração para execução utilizando Docker Compose.

Subir os containers:

```bash
docker compose up -d --build
```

Verificar containers:

```bash
docker ps
```

Visualizar logs:

```bash
docker compose logs -f
```

---

# 🌎 Deploy

A aplicação foi preparada para deploy utilizando:

- Docker Compose
- Backend Go
- Frontend React servido pelo Nginx
- Cloudflare Tunnel para exposição segura

Arquitetura:

```
Usuário
   |
   |
Cloudflare Tunnel
   |
   |
Nginx
   |
   |
---------------------
|                   |
Frontend          Backend
React             Go API
                   |
                 SQLite
```

---

# 🔒 Arquivos ignorados

Arquivos sensíveis não fazem parte do repositório:

```
Backend/.env
Backend/estoque.db

Frontend/node_modules
Frontend/dist
Frontend/.env

data/
```

O banco de dados de produção permanece somente no ambiente do servidor.

---

# 📌 Próximas melhorias

- Melhorar responsividade dos botões de ação na tabela mobile
- Implementar testes automatizados
- Adicionar histórico de movimentações do estoque
- Criar relatórios de consumo
- Migrar SQLite para PostgreSQL em ambientes maiores

---

# 👨‍💻 Autor

**Saulo Rangel**

Projeto desenvolvido como parte do processo de aprendizado em desenvolvimento Full Stack.

Tecnologias estudadas e aplicadas:

- React
- JavaScript
- Go
- APIs REST
- Docker
- Banco de dados
- Deploy em ambiente Linux

---

## 📄 Versão

`v1.0.0`
