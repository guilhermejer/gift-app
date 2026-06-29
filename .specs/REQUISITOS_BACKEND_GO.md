# Requisitos do Backend Go (Data Backend)

## 1. Objetivo
Este backend em Go sera o responsavel por persistencia e leitura de dados.
A API Python (FastAPI) nao acessa banco diretamente e depende deste backend via HTTP.

Base URL atual esperada:
- http://localhost:8080

---

## 2. Fonte de verdade de rotas
- A referencia principal de rotas do backend Go e o swagger em `.docs/swagger.yaml`.
- Este documento detalha os requisitos de integracao para os fluxos de profile/embedding usados pela API Python.

### 2.1 Requisito de embedding
- A API Python deve gerar embeddings usando OpenAI `text-embedding-3-small`.
- O backend Go deve aceitar e persistir o vetor gerado por esse modelo no campo `embedding` da tabela `profiles`.

---

## 3. Modelo de dados alvo

### 3.1 Tabela `profiles`
Campos obrigatorios:
- `friend_id`
- `likes`
- `dislikes`
- `embedding`
- `created_at`
- `updated_at`

Regras:
- Nao existe `profile_id`.
- Chave de relacionamento de `profiles` e `friend_id`.
- `nome` nao pertence a tabela `profiles`.

### 3.2 Tabela `friends`
Campos esperados:
- `friend_id`
- `user_id`
- `user_relation`
- `name`
- `gender`
- `birth_date`
- `city`
- `created_at`
- `updated_at`

Regra:
- Os dados de `friends` sao fornecidos pelo usuario pela interface e persistidos no backend Go.

---

## 4. Contratos de profile (alinhados ao swagger)

### 4.1 GET /friends/{friend_id}/profile
Retorna o profile vinculado ao `friend_id`.

Response esperada (exemplo):
```json
{
  "friendID": "f_123",
  "likes": ["livros", "cafe"],
  "dislikes": ["barulho"],
  "embedding": [0.12, -0.03, 0.45]
}
```

### 4.2 PUT /friends/{friend_id}/profile
Cria ou atualiza profile do friend.

Request esperado:
```json
{
  "friend_id": "f_123",
  "likes": ["livros", "cafe"],
  "dislikes": ["barulho"],
  "embedding": [0.12, -0.03, 0.45]
}
```

Regras obrigatorias:
- `friend_id` deve ser aceito no body.
- `embedding` deve ser aceito como parametro da rota de profile.
- `nome` nao deve ser salvo na tabela `profiles`.

---

## 5. Regra do POST de embedding (API Python)
A API Python expoe `POST /profiles/{friend_id}/embedding` como rota de orquestracao.

Regra obrigatoria do request da API Python:
- O body deve conter `friend_id` obrigatoriamente.

Payload esperado pela API Python:
```json
{
  "friend_id": "f_123",
  "likes": ["livros", "cafe"],
  "dislikes": ["barulho"],
  "name": "Ana",
  "city": "Sao Paulo",
  "user_relation": "amiga"
}
```

Comportamento de integracao:
- A API Python gera o vetor de embedding.
- Depois envia para o backend Go via `PUT /friends/{friend_id}/profile`, incluindo:
  - `friend_id`
  - `embedding`
  - dados complementares de profile quando houver

---

## 6. Codigos HTTP esperados
- 200: leitura/atualizacao bem sucedida
- 201: criacao bem sucedida
- 400: payload invalido
- 404: recurso nao encontrado
- 409: conflito de estado
- 422: validacao semantica
- 500: erro interno

---

## 7. Requisitos de robustez
- Validar payloads recebidos
- Sanitizar textos (tamanho maximo e caracteres invalidos)
- Retornar JSON valido em sucesso e erro
- Logs com request id e latencia
- Timestamps em formato ISO 8601 UTC

---

## 8. Checklist de pronto
- [ ] `PUT /friends/{friend_id}/profile` aceita `friend_id`, `likes`, `dislikes`, `embedding`
- [ ] `GET /friends/{friend_id}/profile` retorna profile por `friend_id`
- [ ] `profiles` persistido sem `nome` e sem `profile_id`
- [ ] `friends` contem os campos de cadastro fornecidos pela interface
- [ ] Integracao com API Python para `POST /profiles/{friend_id}/embedding` funcionando
