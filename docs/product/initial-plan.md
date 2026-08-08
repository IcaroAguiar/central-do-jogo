---
goal: Central do Jogo open source - pesquisa, MVP e validacao publica
version: 1.0
date_created: 2026-07-31
last_updated: 2026-07-31
owner: Icaro Aguiar
status: 'Planned'
tags: [feature, discovery, open-source, go, react, pwa, football]
---

# Introduction

![Status: Planned](https://img.shields.io/badge/status-Planned-blue)

Este plano transforma o brainstorming e o grilling em uma sequencia executavel para uma PWA open source de pre-jogo. A implementacao permanece bloqueada ate a conclusao da pesquisa factual de fontes na Fase 0. O produto deve responder onde assistir, publicar escalacoes oficiais, agregar links de noticias permitidas e enviar alertas essenciais, sem placar ao vivo, IA ou microservicos no MVP.

## 1. Requirements & Constraints

- **REQ-001**: A promessa do MVP e uma central de pre-jogo para clubes suportados, com agenda, transmissoes, escalacoes oficiais e noticias relacionadas.
- **REQ-002**: O MVP suporta o snapshot anual oficial dos clubes da Serie A; promovidos entram apos validacao e rebaixados preservam historico.
- **REQ-003**: A allowlist inicial de pesquisa cobre Serie A, Copa do Brasil, Libertadores e Sul-Americana apenas para clubes suportados; a inclusao final depende da Fase 0.
- **REQ-004**: A agenda oferece filtros de semana, mes e temporada conhecida, mostra somente datas publicadas e explicita data indefinida ou alterada.
- **REQ-005**: A primeira experiencia e busca direta por clubes, apelidos cadastrados e partidas conhecidas, sem IA ou busca semantica.
- **REQ-006**: Visitantes podem definir um clube principal e favoritos localmente; login opcional sincroniza preferencias sem sobrescrita silenciosa.
- **REQ-007**: Cada partida lista todas as transmissoes oficiais conhecidas com canal/plataforma, acesso gratuito ou assinatura, regiao, link oficial quando disponivel, fonte, horario de verificacao e confianca.
- **REQ-008**: Escalacoes estruturam titulares, reservas, tecnico e formacao quando publicados oficialmente, preservando fonte e horario observavel.
- **REQ-009**: A central mostra ate cinco links de noticias deduplicados e de fontes permitidas, priorizando as 72 horas anteriores ao jogo e sem republicar texto integral.
- **REQ-010**: Dados ausentes usam estados distintos: aguardando divulgacao, nao encontrado, divergente e sem cobertura; a ultima tentativa fica visivel.
- **REQ-011**: Alertas Web Push opt-in cobrem transmissao confirmada/alterada e escalacao oficial; a permissao so e pedida depois de seguir um clube.
- **REQ-012**: Alertas sao agrupados por partida, tipo e versao idempotente; oscilacoes nao confirmadas nao notificam.
- **REQ-013**: O painel permite ao mantenedor confirmar, corrigir e marcar divergencias, sempre registrando ator, motivo, instante e antes/depois.
- **REQ-014**: Visitantes podem relatar erro por formulario contextual anonimo com rate limit; relatos nunca alteram dados automaticamente.
- **REQ-015**: Clubes e partidas possuem URLs estaveis, HTML inicial semantico, metadados indexaveis, Web Share e copia de link.
- **REQ-016**: A PWA offline mostra shell e ultimas consultas com rotulo offline, horario do cache e retry; cache nunca parece atual.
- **REQ-017**: OAuth configuravel autentica usuarios e mantenedores; a instancia publica inicial usa Google e nao armazena senhas.
- **REQ-018**: O mantenedor e concedido somente por allowlist segura; primeiro login nunca promove usuario.
- **REQ-019**: Exclusao e exportacao JSON da conta sao autoatendidas; exportacao omite tokens, cookies e endpoints Push.
- **REQ-020**: Analytics first-party usa identificador anonimo local, vincula a conta apenas com login/consentimento e remove identificadores apos 90 dias.
- **REQ-021**: A metrica primaria e retorno as centrais de dois jogos consecutivos do clube principal.
- **REQ-022**: A beta acompanha dois jogos consecutivos de cada clube suportado antes do lancamento publico.
- **REQ-023**: O gate de transmissao exige pelo menos 97% de precisao e 90% de cobertura na beta, com formulas e amostra publicadas.
- **REQ-024**: O gate de escalacao exige 95% de publicacoes corretas em ate cinco minutos da divulgacao oficial observavel.
- **REQ-025**: O SLO controlavel de Push exige 99% das notificacoes aceitas pelo servico Push em ate 60 segundos; entrega ao aparelho nao e prometida.
- **SEC-001**: Rate limits por rota/custo protegem busca, OAuth, Push, relatos e admin; o proxy aplica limites globais.
- **SEC-002**: Segredos ficam fora do repositorio e dos logs; configuracao usa variaveis/segredos do operador com valores de exemplo nao sensiveis.
- **SEC-003**: Exportacao, exclusao, preferencias e admin exigem autorizacao, protecao CSRF quando aplicavel, cookies seguros e respostas sem cache publico.
- **SEC-004**: Cada fonte exige manifesto documentando finalidade, acesso, termos/robots aplicaveis, limites, atribuicao, estabilidade, dados e remocao; nao contornar bloqueios.
- **CON-001**: O projeto deve operar com baixo consumo e sem servicos pagos obrigatorios numa VPS existente.
- **CON-002**: IA, placar ao vivo, eventos, estatisticas e pos-jogo ficam fora do MVP.
- **CON-003**: A arquitetura e monolito modular Go, React + Vite, PostgreSQL, REST + OpenAPI e Docker Compose; sem Redis, Kubernetes, gRPC ou runtime Node no servidor.
- **CON-004**: Scheduler e workers Go usam estado, leases, retries e idempotencia no PostgreSQL; frequencia e adaptativa por tipo de dado e proximidade da partida.
- **CON-005**: A API e interna e documentada; nao ha SLA de compatibilidade para terceiros no MVP.
- **CON-006**: Logs estruturados sao a observabilidade tecnica; saude operacional persiste ultimo sucesso/erro/proxima execucao no PostgreSQL e aparece no painel.
- **CON-007**: O repositorio Apache-2.0 nao distribui escudos; operadores fornecem pacote opcional e o produto tem fallback proprio.
- **CON-008**: Produto e documentacao de usuario usam pt-BR e horario de Brasilia explicito; codigo/contratos/contribuicao tecnica usam ingles, com README de entrada bilingue enxuto.
- **GUD-001**: Disponibilidade mensal alvo de 99,5%; falha de adaptador degrada somente a fonte/dado afetado.
- **GUD-002**: p75 mobile deve manter LCP <= 2,5s, INP <= 200ms e CLS <= 0,1; API de leitura p95 <= 500ms sob carga documentada.
- **GUD-003**: WCAG 2.2 AA bloqueia lancamento para fluxos publicos e administrativos.
- **GUD-004**: Backup PostgreSQL criptografado e semanal; restauracao isolada e verificada trimestralmente, aceitando RPO de ate uma semana.
- **PAT-001**: Todo dado normalizado preserva evidencias imutaveis de fonte, valor observado, instante, parser e execucao.
- **PAT-002**: Confianca e derivada por regras deterministicas de prioridade, atualidade e concordancia; override humano exige justificativa auditada.
- **PAT-003**: Cada adaptador inclui manifesto, fixtures redigidas e testes deterministas; CI nunca depende da fonte ao vivo.
- **PAT-004**: Go renderiza HTML e dados iniciais essenciais; React assume interacoes e capacidades PWA por progressive enhancement.

## 2. Implementation Steps

### Implementation Phase 0 - Source feasibility research

- GOAL-001: Validar se fontes gratuitas e admissiveis sustentam a promessa antes de iniciar a implementacao do produto.

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-001 | Criar `docs/research/source-evaluation-template.md` com campos de SEC-004, cobertura, formato, freshness, rate limit, atribuicao, evidencia e decisao. | x | 2026-07-31 |
| TASK-002 | Criar `docs/research/schedules.md` comparando fontes oficiais/publicas para Serie A, Copa do Brasil, Libertadores e Sul-Americana; registrar amostras redigidas em `docs/research/evidence/schedules/`. | x | 2026-08-01 |
| TASK-003 | Criar `docs/research/broadcasts.md` mapeando fontes de transmissao por competicao/detentor e verificando se canal, acesso, regiao e link podem ser obtidos legal e tecnicamente. | x | 2026-08-01 |
| TASK-004 | Criar `docs/research/lineups.md` avaliando fontes oficiais e latencia observavel para titulares/reservas; registrar ao menos tres casos reais em `docs/research/evidence/lineups/`. | x | 2026-08-01 |
| TASK-005 | Criar `docs/research/news.md` com allowlist inicial, RSS/feed/HTML disponivel, atribuicao e janela de 72 horas. | x | 2026-08-01 |
| TASK-006 | Criar `docs/research/source-matrix.md` com uma linha por dado/competicao e estados `viable`, `conditional` ou `rejected`; declarar a allowlist realmente suportavel. | x | 2026-08-01 |
| TASK-007 | Criar `docs/adr/0001-source-feasibility-gate.md`; bloquear GOAL-002 se transmissoes nao tiverem caminho plausivel para REQ-023 ou escalacoes para REQ-024. | x (accept condicional; ver ADR) | 2026-08-01 |

### Implementation Phase 1 - Runtime and project foundation

- GOAL-002: Criar a fundacao reproduzivel do produto somente depois da aprovacao de GOAL-001.
- Sequencia (2026-07-31): excecao aprovada para scaffold vazio de fundacao antes de GOAL-001 (health, shell PWA, Compose, CI, docs OSS). Dominio, adaptadores e jornadas de produto permanecem bloqueados por DEP-001. Ver `docs/adr/0000-toolchain-versions.md`.

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-008 | Criar `go.mod`, `cmd/server/main.go`, `cmd/worker/main.go` e `internal/platform/config/config.go`; implementar shutdown por contexto e erros explicitos. Depende de TASK-007. | x (scaffold vazio; dependencia de TASK-007 relaxada apenas para fundacao) | 2026-07-31 |
| TASK-009 | Criar `web/package.json`, `web/vite.config.ts`, `web/src/main.tsx`, `web/src/app/App.tsx` e `web/public/manifest.webmanifest` para a PWA React/Vite. | x | 2026-07-31 |
| TASK-010 | Criar `api/openapi.yaml`, `internal/platform/http/router.go` e `web/src/api/generated/`; gerar/validar tipos do contrato REST sem publicar SLA externo. | x (contrato minimo `/healthz`) | 2026-07-31 |
| TASK-011 | Criar `deploy/Dockerfile`, `deploy/compose.yaml`, `.env.example` e `deploy/README.md`; servir o build da PWA sem runtime Node e iniciar PostgreSQL com healthchecks. | x | 2026-07-31 |
| TASK-012 | Criar `LICENSE`, `README.md`, `README.en.md`, `CONTRIBUTING.md`, `SECURITY.md` e `CODE_OF_CONDUCT.md` conforme CON-007 e CON-008. | x | 2026-07-31 |
| TASK-013 | Criar `.github/workflows/ci.yml` com Go format/vet/lint/test/race, frontend lint/typecheck/unit, contrato OpenAPI e testes PostgreSQL/Playwright condicionados por escopo. | x (CI base; e2e publico obrigatorio a partir da ponte `docs/validation/`) | 2026-07-31 |

### Implementation Phase 2 - Domain, persistence and ingestion

- GOAL-003: Implementar o nucleo auditavel de jogos, evidencias, fontes e jobs.

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-014 | Criar `internal/domain/club.go`, `competition.go`, `match.go`, `broadcast.go`, `lineup.go`, `news.go` e `evidence.go` com IDs estaveis, instantes UTC e estados de REQ-010. | x | 2026-08-01 |
| TASK-015 | Criar migracoes em `db/migrations/` para clubes/temporadas/competicoes/partidas, evidencias, transmissoes, escalacoes, noticias, fontes e auditoria; adicionar constraints de integridade. | x | 2026-08-01 |
| TASK-016 | Criar `internal/sources/adapter.go`, `manifest.go`, `registry.go` e `testkit/`; implementar contrato de PAT-003 e rejeitar adaptador sem manifesto valido. | x | 2026-08-01 |
| TASK-017 | Implementar apenas os adaptadores aprovados em `internal/sources/<source_id>/`, cada um com `manifest.yaml`, `adapter.go`, `fixtures/` e `adapter_test.go`. | x | 2026-08-01 |
| TASK-018 | Criar `internal/reconciliation/` para regras deterministicas de PAT-002, estados de divergencia e overrides versionados. | x | 2026-08-01 |
| TASK-019 | Criar `internal/jobs/` e tabelas `jobs`, `job_attempts`, `source_health`; implementar leases PostgreSQL, retries, idempotencia e cadencia adaptativa. | x | 2026-08-01 |
| TASK-020 | Criar `internal/platform/logging/` com JSON estruturado e redacao de segredos; registrar correlation ID, source ID, job ID e match ID sem dados pessoais desnecessarios. | x | 2026-08-01 |

### Implementation Phase 3 - Public PWA journeys

- GOAL-004: Entregar busca, agenda e central de pre-jogo publicas, indexaveis e acessiveis.

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-021 | Implementar `internal/features/search/` e `web/src/features/search/` para clubes/apelidos/partidas, com rate limit e navegacao por teclado. | x | 2026-08-01 |
| TASK-022 | Implementar `internal/features/clubs/` e `web/src/features/clubs/` para agenda semana/mes/temporada, principal/favoritos locais e estados de calendario. | x | 2026-08-01 |
| TASK-023 | Implementar `internal/features/matches/` e `web/src/features/matches/` para transmissoes, confianca, escalacoes, noticias e estados de lacuna. | x | 2026-08-01 |
| TASK-024 | Criar `internal/platform/render/` e templates em `web/server-templates/` para HTML inicial, metadados, canonical URL, Open Graph e dados iniciais serializados com escaping seguro. | x | 2026-08-01 |
| TASK-025 | Implementar service worker em `web/src/pwa/` com app shell, cache de ultimas consultas, timestamp, estado offline e retry; nunca cachear respostas privadas/exportacoes. | x | 2026-08-01 |
| TASK-026 | Implementar Web Share/copy em `web/src/features/sharing/` e fallback acessivel; validar previews de clubes e partidas. | x | 2026-08-01 |

### Implementation Phase 4 - Identity, Push and operations

- GOAL-005: Entregar conta opcional, alertas essenciais e painel do mantenedor.

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-027 | Criar `internal/features/auth/` para OAuth configuravel, Google inicial, sessoes seguras e allowlist de mantenedor; adicionar migracoes de usuarios/sessoes. | x | 2026-08-02 |
| TASK-028 | Criar `internal/features/preferences/` e `web/src/features/preferences/` para merge local/remoto, escolha de clube principal e favoritos. | x | 2026-08-02 |
| TASK-029 | Criar `internal/features/push/` e `web/src/features/push/` para consentimento contextual, subscriptions, deduplicacao/versionamento, retries e limpeza de endpoints expirados. | x | 2026-08-03 |
| TASK-030 | Criar `internal/features/privacy/` e `web/src/features/settings/` para eventos first-party, retencao 90 dias, exportacao JSON e exclusao autoatendida. | | |
| TASK-031 | Criar `internal/features/admin/` e `web/src/features/admin/` para saude de fontes, partidas em risco, confirmacao/correcao/divergencia e trilha de auditoria. | | |
| TASK-032 | Criar `internal/features/reports/` e `web/src/features/reports/` para relatos anonimos limitados/sanitizados e fila contextual do mantenedor. | | |

### Implementation Phase 5 - Hardening, beta and release

- GOAL-006: Provar requisitos de qualidade e liberar apenas quando os gates forem atendidos.

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-033 | Criar `docs/validation/beta-protocol.md` e acompanhar dois jogos consecutivos de cada clube suportado; capturar denominadores, fontes, timestamps, erros e lacunas. | | |
| TASK-034 | Criar `docs/validation/beta-results.md`; calcular REQ-023, REQ-024, REQ-025 e REQ-021 sem excluir falhas conhecidas da amostra. | | |
| TASK-035 | Criar suites Playwright em `e2e/public/` e `e2e/admin/`, incluindo busca, agenda, central, offline, login, preferencias, Push simulado, correcao e relato. | | |
| TASK-036 | Executar auditoria WCAG 2.2 AA e budgets GUD-002; registrar evidencias em `docs/validation/accessibility.md` e `performance.md`. | | |
| TASK-037 | Criar `ops/backup.sh`, `ops/restore-test.sh` e `docs/operations/backup-restore.md`; provar backup criptografado e restauracao isolada sem registrar segredos. | | |
| TASK-038 | Criar `docs/operations/runbook.md` para fonte quebrada, dado divergente, Push falho, OAuth indisponivel, rollback e remocao de adaptador. | | |
| TASK-039 | Liberar a versao `v0.1.0` somente se TASK-034, TASK-036 e TASK-037 passarem; caso contrario manter status beta e publicar bloqueios/residuos. | | |

### Implementation Phase 6 - Post-MVP continuity

- GOAL-007: Expandir cobertura sem enfraquecer a allowlist validada.

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-040 | Repetir GOAL-001 para Serie B antes da temporada posterior; criar `docs/research/serie-b-source-matrix.md` e nao privilegiar apenas clubes rebaixados. | | |
| TASK-041 | Submeter funcoes extras de calendario/contagem, palpites ou outras ideias a novo discovery; nenhuma esta aprovada por este plano. | | |

## 3. Alternatives

- **ALT-001**: Scraping-only foi rejeitado; feeds, JSON e dados estruturados sao preferidos e HTML e usado por adaptadores quando necessario.
- **ALT-002**: Full-stack TypeScript foi rejeitado em favor de Go modular + React/Vite para aprofundar Go e manter PWA rica.
- **ALT-003**: Microservicos, gRPC, Redis e Kubernetes foram rejeitados por custo e complexidade prematuros.
- **ALT-004**: SQLite foi rejeitado em favor de PostgreSQL por concorrencia, auditoria e jobs persistentes.
- **ALT-005**: E-mail e apps nativos foram rejeitados; alertas usam Web Push da PWA.
- **ALT-006**: API publica read-only foi adiada; OpenAPI documenta apenas o contrato interno do MVP.
- **ALT-007**: Placar/eventos ao vivo e pos-jogo foram rejeitados para preservar foco no pre-jogo.
- **ALT-008**: Escudos no repositorio foram rejeitados; pacote opcional e responsabilidade do operador.
- **ALT-009**: Observabilidade com metricas/tracing foi rejeitada; logs estruturados + saude operacional no PostgreSQL foram escolhidos.
- **ALT-010**: Serie B no MVP foi rejeitada; fica na fase seguinte para manter a beta viavel.

## 4. Dependencies

- **DEP-001**: Resultado aprovado de `docs/research/source-matrix.md`; sem ele nenhuma implementacao de produto comeca.
- **DEP-002**: Go e toolchain suportada escolhidos/pinados na Fase 1, com `gofmt`, `go vet`, linter e race detector.
- **DEP-003**: Node/pnpm somente em build/CI do frontend; nenhum runtime Node em producao.
- **DEP-004**: PostgreSQL suportado e volume persistente com backup externo criptografado.
- **DEP-005**: Credenciais Google OAuth, VAPID e segredos do operador fornecidos fora do repositorio.
- **DEP-006**: HTTPS e reverse proxy compativeis com OAuth, secure cookies, PWA e Web Push.
- **DEP-007**: Navegadores com PWA/Web Push; comportamento e limitacoes por plataforma devem ser documentados na Fase 0/1.
- **DEP-008**: Pacote opcional de escudos fornecido/licenciado pelo operador; fallback interno nao depende dele.

## 5. Files

- **FILE-001**: `docs/research/` - pesquisa e matriz de fontes.
- **FILE-002**: `docs/adr/` - decisoes arquiteturais e gates.
- **FILE-003**: `cmd/server/main.go` - servidor HTTP e ciclo de vida.
- **FILE-004**: `cmd/worker/main.go` - scheduler/workers.
- **FILE-005**: `internal/domain/` - modelo esportivo e evidencias.
- **FILE-006**: `internal/sources/` - adaptadores, manifestos e testkit.
- **FILE-007**: `internal/reconciliation/` - confianca e overrides.
- **FILE-008**: `internal/jobs/` - leases, retries e saude.
- **FILE-009**: `internal/features/` - casos de uso publicos, conta, Push e admin.
- **FILE-010**: `internal/platform/` - config, HTTP, render, logging e seguranca transversal.
- **FILE-011**: `db/migrations/` - schema PostgreSQL.
- **FILE-012**: `api/openapi.yaml` - contrato REST interno.
- **FILE-013**: `web/src/` - PWA React/Vite.
- **FILE-014**: `web/server-templates/` - HTML inicial semantico.
- **FILE-015**: `deploy/` - imagem e Docker Compose.
- **FILE-016**: `e2e/` - Playwright das jornadas criticas.
- **FILE-017**: `ops/` - backup e restore test.
- **FILE-018**: `docs/validation/` - protocolo/resultados da beta e gates.
- **FILE-019**: `docs/operations/` - runbooks e continuidade.
- **FILE-020**: `.github/workflows/ci.yml` - gates de merge.

## 6. Testing

- **TEST-001**: Table-driven tests e fuzzing para parsers/normalizadores usando fixtures, sem rede.
- **TEST-002**: Contract tests garantem que todo adaptador produz o mesmo modelo e preserva evidencias.
- **TEST-003**: Testes de reconciliacao cobrem prioridade, atualidade, concordancia, divergencia e override auditado.
- **TEST-004**: Testes PostgreSQL reais cobrem constraints, leases concorrentes, idempotencia, retries e retencao.
- **TEST-005**: Race detector cobre workers, shutdown/cancelamento e processamento concorrente em Go.
- **TEST-006**: OpenAPI valida requests/responses e gera tipos frontend sem drift.
- **TEST-007**: Testes React cobrem busca, estados de lacuna, confianca, preferencias e acessibilidade de componentes.
- **TEST-008**: Playwright publico cobre busca -> clube -> partida -> seguir -> Push e compartilhamento.
- **TEST-009**: Playwright admin cobre allowlist, saude, correcao, auditoria e relato anonimo.
- **TEST-010**: Teste offline prova rotulo/timestamp, ausencia de mutacao e nao-cache de dados privados.
- **TEST-011**: Testes de seguranca cobrem OAuth state/nonce, sessao, CSRF, autorizacao, rate limit e redacao de logs.
- **TEST-012**: Teste de privacidade cobre merge de preferencias, exportacao, exclusao e expurgo aos 90 dias.
- **TEST-013**: Carga documentada valida API p95 <= 500ms e concorrencia-alvo da VPS.
- **TEST-014**: Lighthouse/browser evidence valida Core Web Vitals e WCAG 2.2 AA nas jornadas criticas.
- **TEST-015**: Beta real valida 97%/90% de transmissao, 95%/5min de escalacao e 99%/60s de envio Push.

## 7. Risks & Assumptions

- **RISK-001**: Fontes gratuitas podem nao atingir cobertura/latencia; GOAL-001 pode reduzir a allowlist ou bloquear o produto.
- **RISK-002**: Termos, robots ou layouts podem mudar; cada adaptador precisa falhar fechado e ter remocao rapida.
- **RISK-003**: Direitos de transmissao variam por regiao/plano e podem mudar perto do jogo, afetando REQ-023.
- **RISK-004**: Escalacoes oficiais podem existir apenas em imagem/rede social dificil de estruturar, afetando REQ-024.
- **RISK-005**: Web Push varia por navegador/SO; medir aceite pelo servico, nao prometer entrega final.
- **RISK-006**: Logs-only reduz deteccao agregada; o painel de saude e a disciplina operacional tornam-se criticos.
- **RISK-007**: Backup semanal aceita perda de ate uma semana; a decisao deve permanecer visivel ao operador.
- **RISK-008**: Google OAuth cria dependencia externa; falha do provedor nao pode derrubar conteudo publico.
- **RISK-009**: HTML inicial Go + React pode duplicar markup/estado; contratos de progressive enhancement devem ser testados.
- **RISK-010**: Relatos anonimos podem gerar abuso; rate limit, sanitizacao e retencao precisam ser verificaveis.
- **ASSUMPTION-001**: A VPS existente suporta um app Go, PostgreSQL, reverse proxy e volume da beta dentro do custo atual.
- **ASSUMPTION-002**: O operador consegue fornecer HTTPS, Google OAuth, VAPID e destino externo de backup.
- **ASSUMPTION-003**: A lista oficial de clubes e competicoes pode ser obtida de fonte admissivel e versionada por temporada.
- **ASSUMPTION-004**: O usuario aceita que funcoes divertidas extras permanecam fora do plano ate novo discovery.

## 8. Related Specifications / Further Reading

- Tarefa viva: https://app.notion.com/p/3ad4fec0ff5f812184b9ff3d0ea5dd86
- Contexto original: chatgpt-conversation://6a6b736f-1d88-83e9-8dfb-f8cdda6dc50a
- O primeiro artefato executavel deste plano e `docs/research/source-evaluation-template.md`; nenhum codigo de produto deve ser criado antes de GOAL-001.
- Excecao documentada em 2026-07-31: scaffold vazio de fundacao (TASK-008..013 sem dominio) pode existir antes de GOAL-001; ver `docs/adr/0000-toolchain-versions.md`.
