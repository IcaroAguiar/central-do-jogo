# Central do Jogo — brainstorming e grilling

Status: descoberta em andamento. Este documento separa decisões confirmadas de hipóteses. Não é PRD nem autorização de implementação.

## Oportunidade

Reduzir o esforço de descobrir onde assistir a uma partida e concentrar, em uma única página de pré-jogo, as informações que o torcedor procura imediatamente antes do início.

## Decisões confirmadas

- Promessa do MVP: central do jogo.
- Cobertura inicial: clubes da Série A, incluindo seus jogos em diferentes competições.
- Jornada coberta: pré-jogo até a divulgação da escalação oficial.
- Verdade de transmissão: fontes oficiais e editoriais reputadas, com proveniência, última verificação e estado de confiança.
- Identidade: conteúdo público com conta opcional para sincronizar times e preferências.
- Coleta: feeds, JSON e dados estruturados quando disponíveis; scraping HTML por adaptadores isolados quando necessário.
- Correções: painel mínimo de mantenedor, com histórico e autoria.
- Alertas: transmissão confirmada/alterada e escalação oficial publicada.
- Superfície: web responsiva PWA, distribuída por URL e instalável.
- Fora do MVP: placar, eventos ao vivo, estatísticas e pós-jogo.
- IA não é requisito.

## Ideias por perspectiva

### Produto

1. Cartão de partida com todos os meios oficiais de transmissão e estado de confiança.
2. Página pública de cada clube com agenda de semana, mês e temporada.
3. Conta opcional para seguir clubes e sincronizar preferências.
4. Alertas essenciais de transmissão e escalação.
5. Histórico de correções e divergências para transformar confiabilidade em diferencial visível.

### Design

1. Resposta "onde assistir" acima da dobra, sem texto intermediário.
2. Linha do tempo de pré-jogo: agendado, transmissão apurada, escalação aguardada e escalação oficial.
3. Selo de confiança com fonte e horário de verificação, explicado em linguagem simples.
4. Seletor rápido de clube e período, com retorno imediato à agenda personalizada.
5. Bloco compacto de notícias externas com título, veículo, horário e link, sem republicar conteúdo.

### Engenharia

1. Adaptadores isolados por fonte, com contrato comum de normalização.
2. Modelo de evidência que preserve fonte, instante da coleta, valor observado e confiança calculada.
3. Pipeline com frequências diferentes: agenda lenta, transmissão crescente perto do jogo e escalação intensiva no pré-jogo.
4. Painel de operações para resolver entidades, confirmar dados e registrar overrides auditáveis.
5. Monitor de saúde por fonte, com detecção de parser quebrado e dados envelhecidos.

## Top 5 preliminar

1. Resposta imediata de transmissão com proveniência.
   - Motivo: entrega a dor original e diferencia pela transparência.
   - Hipótese: torcedores preferem uma resposta curta com confiança explícita a uma lista editorial longa.
2. Página de pré-jogo com linha do tempo até a escalação.
   - Motivo: dá coesão à promessa "central do jogo" sem entrar em dados ao vivo.
   - Hipótese: transmissão + escalação são suficientes para criar retorno no dia da partida.
3. Adaptadores de fonte e evidência normalizada.
   - Motivo: concentra a principal complexidade técnica e permite substituir fontes.
   - Hipótese: fontes gratuitas e públicas cobrem os clubes da Série A com qualidade aceitável.
4. Correção operacional auditável.
   - Motivo: a promessa de confiabilidade exige uma rota humana quando fontes falham.
   - Hipótese: o volume de exceções permanece administrável por poucos mantenedores.
5. Alertas essenciais.
   - Motivo: transforma consulta passiva em utilidade recorrente.
   - Hipótese: usuários aceitam opt-in para dois eventos de alto valor e baixo ruído.

## Próximas decisões após a superfície

1. Canal de entrega dos alertas.
2. Frequência e janela de atualização por tipo de dado.
3. definição exata de "notícia relacionada" e critérios editoriais.
4. fonte/cobertura mínima que permite declarar um clube ou competição suportado.
5. arquitetura inicial em Go, armazenamento e agendador.
6. métricas de sucesso e critérios de saída do MVP.
7. política de scraping, robots.txt, atribuição, cache e remoção de fontes.
8. licença, governança e contribuição open source.
