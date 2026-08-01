# Documentação do projeto

Este repositório é público. A documentação versionada deve ajudar usuários, contribuidores e mantenedores a compreender o produto e reproduzir suas decisões.

## Documentos públicos

- `product/`: visão, requisitos, escopo e notas de discovery revisadas.
- `research/`: pesquisas sintetizadas, matrizes de fontes e evidências redigidas.
- `adr/`: decisões arquiteturais e seus trade-offs (inclui pinagem de toolchain em `0000-toolchain-versions.md`).

Antes de publicar, remova credenciais, dados pessoais, URLs privadas, conteúdo de terceiros sem licença e evidências brutas que não possam ser redistribuídas.

## Documentos não versionados

Use `docs/private/` ou `docs/internal/` para rascunhos pessoais e evidências que não devem ser públicas. Arquivos em qualquer pasta também podem usar os sufixos `*.private.md` ou `*.local.md`. Esses caminhos são ignorados pelo Git.

O `.gitignore` evita commits acidentais, mas não é um cofre. Segredos, tokens, cookies e credenciais não devem ser gravados nem mesmo nas pastas ignoradas.
