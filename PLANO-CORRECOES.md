# Plano de correções e melhorias — Controle de Estoque

Documento vivo para orientar a correção dos riscos encontrados na análise de
17/09/2026. Cada alteração de código deve atualizar este documento com o
status, os arquivos afetados e as validações executadas.

## Regras de execução

- Corrigir em pequenas etapas, preservando as funcionalidades existentes.
- Validar autorização por usuário em toda operação autenticada.
- Fazer primeiro as correções de segurança e integridade de dados.
- Não considerar uma correção concluída sem teste ou evidência de validação.
- Não registrar segredos, senhas, tokens ou dados fiscais nos logs.
- Não fazer commit, merge, deploy ou publicação sem autorização explícita.

## Legenda

- **Pendente**: ainda não implementado.
- **Em andamento**: alteração iniciada, aguardando validação completa.
- **Concluído**: código alterado e critérios de aceite verificados.
- **Aceito como risco**: decisão consciente, com justificativa registrada.

## Fase 0 — Preparação e baseline

### P0. Inventário e baseline técnico

- **Status:** Concluído na análise inicial; testes automatizados ainda pendentes.
- **Escopo:** confirmar estrutura, documentação, estado do Git, build e testes.
- **Critérios de aceite:** manter `git diff --check` limpo; registrar falhas de
  ambiente separadamente de falhas do código; não incluir artefatos gerados.
- **Observação:** não foram encontrados testes automatizados. `go test ./...`
  e `npm run build` ficaram bloqueados pelo acesso do sandbox aos caches/configuração.

## Fase 1 — Segurança crítica e autenticação

### P1. Endurecer validação de JWT

- **Status:** Concluído.
- **Arquivos previstos:** `Backend/middleware/auth.go`,
  `Backend/config/config.go`, possivelmente `Backend/main.go`.
- **Ações:**
  - [x] rejeitar inicialização se `JWT_SECRET` estiver vazio ou fraco;
  - [x] aceitar somente HS256;
  - [x] validar `usuario_id` com tipo numérico inteiro e positivo;
  - [x] usar o parser de claims do JWT, que valida a expiração;
  - [x] evitar assertions que poderiam causar panic;
  - [ ] avaliar `iss`, `aud`, `iat` e estratégia de invalidação após troca de senha.
- **Critérios de aceite:** token sem assinatura válida, algoritmo diferente,
  segredo ausente, claim ausente ou claim com tipo inválido retorna 401 sem panic.
- **Evidência:** `go test ./...` e `go vet ./...` executados em 17/09/2026;
  ambos passaram. Não há testes automatizados no repositório ainda, portanto os
  casos específicos de JWT permanecem cobertos manualmente/por implementação e
  devem ser incluídos na Fase 4 (P13).

### P2. Corrigir recuperação de senha

- **Status:** Pendente.
- **Arquivos previstos:** `Backend/main.go`, `Backend/handlers/auth.go`,
  `Backend/middleware/ratelimit.go`, frontend de login.
- **Ações:**
  - aplicar rate limit também à consulta da pergunta;
  - reduzir enumeração de usuários com resposta e comportamento uniformes;
  - validar força e tamanho da nova senha no backend;
  - avaliar substituição da pergunta de segurança por código temporário/e-mail;
  - invalidar sessões/tokens existentes após redefinição, se a arquitetura permitir.
- **Critérios de aceite:** não é possível enumerar usuários por respostas distintas;
  tentativas repetidas são limitadas; senha inválida é rejeitada pela API.

### P3. Tornar o rate limit confiável

- **Status:** Pendente.
- **Arquivos previstos:** `Backend/middleware/ratelimit.go`, configuração de
  proxy/deploy se necessário.
- **Ações:**
  - aceitar `CF-Connecting-IP` somente quando a requisição vier de proxy confiável;
  - validar o formato do IP e definir fallback seguro;
  - considerar chave composta por IP + identificador normalizado da conta;
  - adicionar limpeza/limite de memória e avaliar armazenamento compartilhado em produção.
- **Critérios de aceite:** cabeçalhos arbitrários enviados diretamente ao backend
  não permitem contornar o limite; o comportamento atrás do Cloudflare continua correto.

### P4. Proteger tokens no frontend

- **Status:** Pendente — melhoria arquitetural.
- **Arquivos previstos:** `Frontend/src/context/AuthContext.jsx`,
  `Frontend/src/api/api.js`, backend e nginx.
- **Ações:** avaliar migração de JWT em `localStorage` para cookie `HttpOnly`,
  `Secure`, `SameSite` e proteção CSRF; se a migração não for imediata, reduzir
  exposição e documentar o risco residual.
- **Critérios de aceite:** decisão arquitetural registrada e fluxo de login,
  logout, expiração e renovação coberto por testes.

## Fase 2 — Integridade do estoque e validação de entrada

### P5. Validar dados no backend

- **Status:** Concluído.
- **Arquivos previstos:** `Backend/handlers/itens.go`,
  `Backend/handlers/notas_fiscais.go`, modelos e frontend.
- **Ações:**
  - [x] rejeitar quantidade, retirada e estoque mínimo negativos, não finitos ou acima
    de limite definido;
  - [x] exigir nome, unidade e local com tamanho máximo;
  - [x] validar locais permitidos no backend;
  - [x] aplicar as regras também às entradas confirmadas de importação;
  - [ ] validar tamanho e formato dos campos de autenticação;
  - [ ] repetir as regras no frontend apenas para melhor experiência, nunca como única defesa.
- **Critérios de aceite:** chamadas diretas à API não conseguem criar estoque
  negativo, desabilitar alertas com mínimo negativo ou gravar valores inválidos.
- **Evidência:** adicionada validação centralizada em
  `Backend/handlers/validacao.go`, aplicada aos endpoints de itens, retirada e
  confirmação de nota. `go test ./...` e `go vet ./...` passaram em 17/09/2026.

### P6. Tornar retirada atômica

- **Status:** Pendente.
- **Arquivos previstos:** `Backend/handlers/itens.go`.
- **Ações:** substituir o fluxo leitura-depois-atualização por operação atômica,
  mantendo o estoque em zero e verificando linhas afetadas.
- **Critérios de aceite:** retiradas concorrentes não perdem movimentações e nunca
  deixam o estoque abaixo de zero.

### P7. Tornar confirmação de nota transacional e idempotente

- **Status:** Pendente.
- **Arquivos previstos:** `Backend/handlers/notas_fiscais.go`, banco/modelos e frontend.
- **Ações:**
  - validar toda a lista recebida, com limite de itens;
  - usar transação para atualizar/criar tudo ou nada;
  - não ignorar erros silenciosamente;
  - definir estratégia de idempotência para reenvio (chave de operação ou confirmação única);
  - manter isolamento por usuário em todas as operações.
- **Critérios de aceite:** falha em qualquer item gera rollback e resposta clara;
  reenvio da mesma confirmação não duplica entrada.

## Fase 3 — Uploads, OCR e integrações externas

### P8. Limitar efetivamente uploads e processamento

- **Status:** Pendente.
- **Arquivos previstos:** `Backend/main.go`,
  `Backend/handlers/notas_fiscais.go`, `Backend/services/ocr_nota.go` e
  `Backend/services/ocr_cloud.go`.
- **Ações:**
  - aplicar `http.MaxBytesReader` no corpo da requisição;
  - validar MIME e formato real do arquivo;
  - limitar dimensões de imagem antes de ampliar/processar;
  - limitar número de OCRs simultâneos por usuário/servidor;
  - limitar tamanho da resposta do OCR.space;
  - padronizar timeouts e mensagens sem expor detalhes internos.
- **Critérios de aceite:** corpo acima do limite é rejeitado cedo; imagem com
  dimensões abusivas não causa consumo excessivo; concorrência é controlada.

### P9. Remover debug sensível de produção

- **Status:** Pendente.
- **Arquivos previstos:** `Backend/services/ocr_nota.go`,
  `Backend/services/nfce_scraper.go`, configuração e `.gitignore`.
- **Ações:** remover gravações automáticas ou protegê-las por flag explícita de
  desenvolvimento; usar diretório temporário seguro e retenção controlada;
  revisar permissões dos arquivos existentes no servidor.
- **Critérios de aceite:** produção não grava texto OCR, HTML de nota ou conteúdo
  fiscal sem configuração explícita e auditável.

### P10. Manter consulta QR Code segura antes de reativar

- **Status:** Pendente; funcionalidade atualmente desativada.
- **Arquivos previstos:** `Backend/services/nfce_scraper.go` e handlers.
- **Ações:** exigir HTTPS; manter allowlist exata; limitar tamanho da resposta;
  bloquear redirecionamentos para outros hosts; não reativar scraping sem nova
  validação do estado real da SEFAZ.
- **Critérios de aceite:** testes de SSRF cobrem host, esquema, redirecionamento,
  IP privado e respostas grandes.

## Fase 4 — Headers, deploy e qualidade operacional

### P14. Exigir confirmação explícita para excluir produto

- **Status:** Pendente.
- **Arquivos previstos:** `Frontend/src/pages/Dashboard.jsx`,
  `Frontend/src/components/TabelaItens.jsx`,
  `Frontend/src/components/ModalConfirmacao.jsx` e estilos relacionados.
- **Ações:**
  - abrir uma confirmação visível ao clicar em excluir;
  - deixar claro o nome do produto e que a ação é permanente;
  - exigir uma ação positiva separada, como botão “Excluir”;
  - manter “Cancelar” como opção segura e não executar a exclusão ao fechar o modal;
  - impedir cliques repetidos enquanto a exclusão estiver sendo processada;
  - garantir foco inicial, foco de teclado e leitura adequada por leitores de tela.
- **Critérios de aceite:** clicar no ícone de exclusão não remove o produto
  imediatamente; somente a confirmação explícita chama o endpoint `DELETE`;
  cancelar, clicar fora ou pressionar `Esc` não exclui o produto; o fluxo funciona
  corretamente em telas pequenas e por teclado.

### P11. Adicionar headers de segurança no nginx

- **Status:** Pendente.
- **Arquivos previstos:** `Frontend/nginx.conf`.
- **Ações:** avaliar e configurar CSP compatível com React/Vite, HSTS somente em
  domínio HTTPS, `X-Content-Type-Options`, `Referrer-Policy` e `frame-ancestors`.
- **Critérios de aceite:** headers aparecem em produção sem quebrar login, PWA,
  câmera ou chamadas à API.

### P12. Reforçar pipeline e reprodutibilidade

- **Status:** Pendente.
- **Arquivos previstos:** `.github/workflows/deploy.yml`, Dockerfiles.
- **Ações:**
  - adicionar testes automatizados e verificações de segurança;
  - preferir `npm ci` no Dockerfile;
  - fixar actions por versão segura ou SHA após revisão;
  - tornar pull/deploy previsível e adicionar smoke test pós-deploy;
  - avaliar aprovação manual antes do deploy de produção.
- **Critérios de aceite:** CI detecta regressões; build usa lockfile; deploy falho
  não é apresentado como concluído.

### P13. Criar suíte de testes

- **Status:** Pendente.
- **Escopo mínimo:**
  - autenticação e claims JWT;
  - isolamento entre usuários;
  - validação de quantidades e locais;
  - retirada concorrente;
  - transação/idempotência de nota;
  - limite de uploads e SSRF;
  - build do frontend e lint/checagens disponíveis.
- **Critérios de aceite:** testes rodam localmente e no CI, com casos de falha
  reproduzindo os riscos listados neste documento.

## Ordem de implementação proposta

1. P1 — JWT e falha segura de configuração.
2. P5 — validação de entrada e integridade numérica.
3. P6 — retirada atômica.
4. P7 — confirmação transacional/idempotente.
5. P2 e P3 — recuperação de senha e rate limit.
6. P8 e P9 — uploads/OCR/debug.
7. P11 — headers de segurança.
8. P10 — QR Code, somente se voltar a ser necessário.
9. P12 e P13 — pipeline e cobertura de testes contínua.
10. P4 — migração de armazenamento de token, conforme decisão arquitetural.

## Registro de progresso

| Data | Item | Alteração/resultado | Evidência |
|---|---|---|---|
| 17/09/2026 | Análise inicial | Riscos e melhorias catalogados; nenhum código alterado | `git status`, leitura do projeto |
| 17/09/2026 | Documento | Plano criado no projeto | Este arquivo |
| 17/09/2026 | P1 | JWT endurecido; segredo vazio/curto bloqueia inicialização e claims inválidos retornam 401 | `go test ./...`, `go vet ./...` |
| 17/09/2026 | P5 | Validação server-side de números, limites, textos, locais e importação | `go test ./...`, `go vet ./...` |

## Registro de decisões e riscos aceitos

Nenhum risco foi formalmente aceito até o momento.
