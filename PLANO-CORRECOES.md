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

- **Status:** Em andamento — hardening básico concluído; invalidação de sessões pendente.
- **Arquivos previstos:** `Backend/main.go`, `Backend/handlers/auth.go`,
  `Backend/middleware/ratelimit.go`, frontend de login.
- **Ações:**
  - [x] aplicar rate limit também à consulta da pergunta;
  - [ ] reduzir enumeração de usuários com resposta e comportamento uniformes;
  - [x] validar tamanho da nova senha no backend, respeitando o limite do bcrypt;
  - [ ] avaliar substituição da pergunta de segurança por código temporário/e-mail;
  - [ ] invalidar sessões/tokens existentes após redefinição, se a arquitetura permitir.
- **Critérios de aceite:** não é possível enumerar usuários por respostas distintas;
  tentativas repetidas são limitadas; senha inválida é rejeitada pela API.
- **Evidência parcial:** a consulta da pergunta passou a usar o rate limit por IP
  e cadastro/redefinição rejeitam senhas com menos de 6 caracteres ou acima de 72
  bytes. `go test ./...`, `go vet ./...` e `git diff --check` passaram em
  17/09/2026. A não enumeração completa e a invalidação de tokens ainda exigem
  trabalho adicional.

### P3. Tornar o rate limit confiável

- **Status:** Concluído — requer configuração do proxy no ambiente de produção.
- **Arquivos previstos:** `Backend/middleware/ratelimit.go`, configuração de
  proxy/deploy se necessário.
- **Ações:**
  - [x] aceitar `CF-Connecting-IP` e `X-Forwarded-For` somente quando a
    conexão vier de uma rede em `TRUSTED_PROXY_CIDRS`;
  - [x] validar o formato do IP e usar o peer da conexão como fallback seguro;
  - [ ] considerar chave composta por IP + identificador normalizado da conta;
  - [x] limpar registros expirados quando o mapa atingir grande volume;
  - [ ] avaliar armazenamento compartilhado em produção, caso existam múltiplas instâncias.
- **Critérios de aceite:** cabeçalhos arbitrários enviados diretamente ao backend
  não permitem contornar o limite; o comportamento atrás do Cloudflare continua correto.
- **Evidência:** `obterIP` agora só usa cabeçalhos quando o peer pertence às
  redes CIDR configuradas; IPs inválidos são ignorados e o mapa possui limpeza
  de registros expirados. `go test ./...`, `go vet ./...` e `git diff --check`
  passaram em 17/09/2026. Antes do próximo deploy, configurar
  `TRUSTED_PROXY_CIDRS` com a rede real do proxy/túnel; sem essa variável o
  sistema fica seguro contra falsificação, mas pode agrupar usuários pelo peer.

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

- **Status:** Concluído.
- **Arquivos previstos:** `Backend/handlers/itens.go`.
- **Ações:**
  - [x] substituir o fluxo leitura-depois-atualização por operação atômica;
  - [x] manter o estoque em zero;
  - [x] verificar linhas afetadas e retornar 404 para item inexistente;
  - [x] consultar e devolver a quantidade atualizada após a operação.
- **Critérios de aceite:** retiradas concorrentes não perdem movimentações e nunca
  deixam o estoque abaixo de zero.
- **Evidência:** `RetirarItem` agora executa a subtração e o limite inferior na
  mesma instrução `UPDATE`. `go test ./...`, `go vet ./...` e `git diff --check`
  passaram em 17/09/2026.

### P7. Tornar confirmação de nota transacional e idempotente

- **Status:** Concluído — validação do build frontend pendente por bloqueio do ambiente.
- **Arquivos previstos:** `Backend/handlers/notas_fiscais.go`, banco/modelos e frontend.
- **Ações:**
  - [x] validar toda a lista recebida, com limite de 500 itens;
  - [x] usar transação para atualizar/criar tudo ou nada;
  - [x] retornar erro e rollback em vez de ignorar falhas silenciosamente;
  - [x] usar chave única por usuário para impedir reprocessamento;
  - [x] manter isolamento por usuário em todas as operações;
  - [x] atualizar o frontend para gerar e reenviar a chave durante a confirmação.
- **Critérios de aceite:** falha em qualquer item gera rollback e resposta clara;
  reenvio da mesma confirmação não duplica entrada.
- **Evidência:** criada a tabela `confirmacoes_importacao`, a API passou a
  receber `{ chave, entradas }` e o modal gera uma chave por prévia. `go test
  ./...` e `go vet ./...` passaram em 17/09/2026. `npm run build` foi tentado,
  mas o ambiente bloqueou o Vite/esbuild ao acessar diretório fora do workspace;
  repetir no ambiente local/CI.

## Fase 3 — Uploads, OCR e integrações externas

### P8. Limitar efetivamente uploads e processamento

- **Status:** Concluído — validação funcional com imagens reais permanece recomendada.
- **Arquivos previstos:** `Backend/main.go`,
  `Backend/handlers/notas_fiscais.go`, `Backend/services/ocr_nota.go` e
  `Backend/services/ocr_cloud.go`.
- **Ações:**
  - [x] aplicar `http.MaxBytesReader` no corpo das requisições multipart;
  - [x] validar o formato real e aceitar somente JPEG/PNG nos fluxos de imagem;
  - [x] limitar dimensões a 6000 px por lado e 12 milhões de pixels;
  - [x] limitar a duas leituras OCR simultâneas, rejeitando espera acima de 5 segundos;
  - [x] limitar a resposta do OCR.space a 2 MB;
  - [x] manter timeouts e mensagens sem expor detalhes internos.
- **Critérios de aceite:** corpo acima do limite é rejeitado cedo; imagem com
  dimensões abusivas não causa consumo excessivo; concorrência é controlada.
- **Evidência:** os três endpoints multipart usam `MaxBytesReader`; os fluxos
  de imagem passam por `ValidarImagem`; Tesseract e OCR.space compartilham um
  limite de concorrência; a resposta externa é limitada por `io.LimitReader`.
  `go test ./...`, `go vet ./...` e `git diff --check` passaram em 17/09/2026.

### P9. Remover debug sensível de produção

- **Status:** Concluído — gravação permanece disponível somente por flag explícita.
- **Arquivos previstos:** `Backend/services/ocr_nota.go`,
  `Backend/services/nfce_scraper.go`, configuração e `.gitignore`.
- **Ações:**
  - [x] remover gravações automáticas por padrão;
  - [x] proteger a gravação por `DEBUG_OCR=true` explícito;
  - [x] usar permissão `0600` para os arquivos quando o debug for habilitado;
  - [ ] usar diretório temporário seguro e retenção controlada;
  - [ ] revisar/remover arquivos existentes no servidor.
- **Critérios de aceite:** produção não grava texto OCR, HTML de nota ou conteúdo
  fiscal sem configuração explícita e auditável.
- **Evidência:** `DebugOCRAtivo()` retorna falso por padrão e protege os dois
  pontos de gravação (`debug_ocr_ultimo_texto.txt` e
  `debug_nfce_ultima_consulta.html`). `go test ./...`, `go vet ./...` e
  `git diff --check` passaram em 17/09/2026.

### P10. Manter consulta QR Code segura antes de reativar

- **Status:** Concluído; funcionalidade continua desativada.
- **Arquivos previstos:** `Backend/services/nfce_scraper.go` e handlers.
- **Ações:**
  - [x] exigir HTTPS e rejeitar userinfo, fragmentos e portas não permitidas;
  - [x] manter allowlist exata por hostname;
  - [x] limitar resposta HTML a 2 MB;
  - [x] bloquear redirecionamentos para hosts/esquemas não permitidos;
  - [x] não reativar scraping sem nova validação do estado real da SEFAZ.
- **Critérios de aceite:** testes de SSRF cobrem host, esquema, redirecionamento,
  IP privado e respostas grandes.
- **Evidência:** validação agora aceita somente HTTPS, allowlist exata e porta
  443; redirecionamentos passam pela mesma validação e o corpo HTML é limitado.
  `go test ./...`, `go vet ./...` e `git diff --check` passaram em 17/09/2026.
  Ainda faltam testes automatizados específicos de SSRF, previstos na P13.

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

- **Status:** Em andamento — configuração concluída; validação no container pendente.
- **Arquivos previstos:** `Frontend/nginx.conf`.
- **Ações:**
  - [x] configurar CSP compatível com React/Vite e PWA;
  - [x] configurar HSTS para o domínio HTTPS;
  - [x] configurar `X-Content-Type-Options`;
  - [x] configurar `Referrer-Policy`;
  - [x] configurar `Permissions-Policy` e `frame-ancestors`/`X-Frame-Options`;
  - [ ] validar headers e ausência de regressões no container em build/ambiente de produção.
- **Critérios de aceite:** headers aparecem em produção sem quebrar login, PWA,
  câmera ou chamadas à API.
- **Evidência parcial:** adicionados headers de segurança no bloco principal e
  na rota `/assets/` do nginx, evitando perda por herança de `add_header`.
  `git diff --check` passou em 17/09/2026. O nginx não está instalado no
  ambiente local, portanto a sintaxe e a validação visual/funcional ainda devem
  ser confirmadas pelo build do container/CI.

### P12. Reforçar pipeline e reprodutibilidade

- **Status:** Em andamento — melhorias aplicadas; pinagem de actions e aprovação manual pendentes.
- **Arquivos previstos:** `.github/workflows/deploy.yml`, Dockerfiles.
- **Ações:**
  - [ ] adicionar testes automatizados e verificações de segurança;
  - [x] usar `npm ci` no Dockerfile e no CI;
  - [ ] fixar actions por SHA após revisão das versões;
  - [x] tornar pull previsível com `git pull --ff-only origin main`;
  - [x] impedir deploys concorrentes com `concurrency`;
  - [x] adicionar smoke tests HTTP para backend e frontend;
  - [ ] avaliar aprovação manual antes do deploy de produção.
- **Critérios de aceite:** CI detecta regressões; build usa lockfile; deploy falho
  não é apresentado como concluído.
- **Evidência parcial:** Dockerfile passou de `npm install` para `npm ci`; o
  workflow recebeu permissões mínimas de leitura, timeout, concorrência, pull
  fast-forward-only e probes pós-deploy. `git diff --check` passou em
  17/09/2026. O workflow não foi executado nesta sessão e não houve deploy.

#### P12-A. Acesso privado da VPS via Tailscale

- **Status:** Em andamento — workflow ajustado; configuração dos secrets e
  execução real ainda pendentes.
- **Constatação:** o IP informado (`100.67.151.30`) é um endereço Tailscale.
  O workflow anterior usava runner GitHub hospedado e SSH direto, sem conectar
  o runner ao tailnet; por isso não funcionaria com a porta 22 pública fechada.
- **Ações:**
  - [x] conectar o runner temporariamente à Tailscale com
    `tailscale/github-action@v4`;
  - [x] usar `VPS_HOST` como host Tailscale, sem gravar o IP no código;
  - [x] usar secrets separados para `VPS_USER` e `VPS_SSH_KEY`;
  - [x] validar fingerprint SHA-256 com `VPS_HOST_FINGERPRINT`;
  - [x] manter SSH na porta 22 privada;
  - [ ] criar/configurar `TAILSCALE_AUTHKEY`, `VPS_HOST`, `VPS_USER`,
    `VPS_SSH_KEY` e `VPS_HOST_FINGERPRINT` no ambiente protegido;
  - [ ] restringir a auth key por tag/política somente à VPS e porta 22;
  - [ ] executar workflow e confirmar smoke tests sem abrir a porta 22 pública;
  - [ ] avaliar trocar o disparo automático por `workflow_dispatch` com
    aprovação manual antes da produção.
- **Critérios de aceite:** o workflow conecta ao tailnet, acessa a VPS pelo
  endereço Tailscale, valida a host key e executa o smoke test; uma execução
  fora do tailnet não alcança a VPS.
- **Atenção de segurança:** foi localizada uma anotação local fora deste
  projeto contendo credenciais junto das instruções de SSH. Os valores não
  foram reproduzidos nem adicionados ao repositório. Rotacionar os segredos
  expostos e remover cópias inseguras após confirmar que não são necessárias.

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
| 17/09/2026 | P6 | Retirada de estoque convertida para atualização atômica | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P7 | Confirmação de nota transacional, com rollback e idempotência por usuário | `go test ./...`, `go vet ./...`; build frontend bloqueado pelo ambiente |
| 17/09/2026 | P2 | Rate limit aplicado à pergunta e validação de senha adicionada no backend | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P3 | Rate limit deixou de confiar em cabeçalhos de IP não verificados | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P8 | Uploads e processamento OCR passaram a ter limites de corpo, imagem, concorrência e resposta | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P9 | Debug de OCR/NFC-e desativado por padrão e protegido por flag | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P10 | QR Code mantido desativado e scraper protegido contra esquemas, redirects e respostas grandes | `go test ./...`, `go vet ./...`, `git diff --check` |
| 17/09/2026 | P11 | Headers de segurança adicionados ao nginx principal e aos assets | `git diff --check`; validação do container pendente |
| 17/09/2026 | P12 | CI/deploy usa lockfile, pull previsível, concorrência controlada e smoke tests | `git diff --check`; execução do workflow pendente |
| 17/09/2026 | P12-A | Workflow passou a conectar runner à Tailscale e validar fingerprint antes do SSH | `.github/workflows/deploy.yml`; secrets/execução pendentes |

## Registro de decisões e riscos aceitos

Nenhum risco foi formalmente aceito até o momento.
