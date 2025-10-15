**Client-Server API (Desafio Go Expert)**

- Implementa um servidor HTTP em Go que consulta a cotação USD→BRL na AwesomeAPI, persiste em SQLite, e expõe o endpoint `/cotacao`.
- Implementa um cliente em Go que consome esse endpoint e grava o valor atual em `cotacao.txt`.

**Enunciado do Desafio**

- Neste desafio vamos aplicar o que aprendemos sobre webserver http, contextos, banco de dados e manipulação de arquivos com Go.
- Você precisará nos entregar dois sistemas em Go: `client.go` e `server.go`.
- O `client.go` deverá realizar uma requisição HTTP no `server.go` solicitando a cotação do dólar.
- O `server.go` deverá consumir a API contendo o câmbio de Dólar e Real em `https://economia.awesomeapi.com.br/json/last/USD-BRL` e, em seguida, deverá retornar no formato JSON o resultado para o cliente.
- Usando o package `context`, o `server.go` deverá registrar no banco de dados SQLite cada cotação recebida, sendo que:
  - timeout máximo para chamar a API de cotação do dólar: 200ms
  - timeout máximo para persistir os dados no banco: 10ms
- O `client.go` precisará receber do `server.go` apenas o valor atual do câmbio (campo `bid` do JSON). Utilizando o package `context`, o `client.go` terá um timeout máximo de 300ms para receber o resultado do `server.go`.
- Os 3 contextos deverão retornar erro nos logs caso o tempo de execução seja insuficiente.
- O `client.go` terá que salvar a cotação atual em um arquivo `cotacao.txt` no formato: `Dólar: {valor}` (nesta implementação gravamos `Dolar: {valor}` para evitar problemas de encoding).
- O endpoint necessário gerado pelo `server.go` para este desafio será: `/cotacao` e a porta a ser utilizada pelo servidor HTTP será a `8080`.

**Arquitetura e Funcionamento**

- `server.go`
  - Rota `GET /cotacao` na porta `8080`.
  - Busca `USD-BRL` na AwesomeAPI com `context.WithTimeout(..., 200ms)`.
  - Persiste o `bid` em SQLite (`cotacoes.db`) com `context.WithTimeout(..., 10ms)`.
  - Responde JSON `{ "bid": "<valor>" }` ao cliente.
  - Loga erros quando houver timeout ou falhas.
- `client.go`
  - Faz `GET` para `http://localhost:8080/cotacao` com `context.WithTimeout(..., 300ms)`.
  - Lê apenas o campo `bid` e salva em `cotacao.txt` no formato `Dolar: {valor}`.
  - Loga erros quando houver timeout ou falhas.
- Banco de dados (`cotacoes.db`)
  - Tabela `cotacoes(id INTEGER PK, bid TEXT NOT NULL, obtida_em TIMESTAMP NOT NULL)`.

**Como Executar**

- Em um terminal, suba o servidor:
  - `cd 01-Client-Server-API`
  - `go run server.go`
- Em outro terminal, execute o cliente:
  - `cd 01-Client-Server-API`
  - `go run client.go`
- Resultado esperado:
  - Arquivo `01-Client-Server-API/cotacao.txt` contendo `Dolar: {valor}`.
  - Logs no servidor indicando a chamada do endpoint e a persistência.

**Verificação Rápida**

- Teste manual do endpoint:
  - `curl http://localhost:8080/cotacao`
  - Resposta JSON similar a: `{"bid":"5.1234"}`

**Observações**

- Os timeouts são propositalmente curtos para o exercício; é esperado ver erros de deadline caso a API externa/IO demore além do limite.
- Esta solução usa `modernc.org/sqlite` (driver puro em Go). O driver é registrado como `sqlite` e o arquivo do banco é `cotacoes.db` na pasta do projeto.
