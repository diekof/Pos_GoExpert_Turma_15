// client.go
package main

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

const (
    serverURL     = "http://localhost:8080/cotacao"
    clientTimeout = 300 * time.Millisecond // timeout do cliente
    outFile       = "cotacao.txt"
)

type cotacaoOut struct {
    Bid string `json:"bid"`
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
    if err != nil {
        log.Fatalf("erro criando request: %v", err)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        // se deadline estourar, cai aqui com context deadline exceeded
        log.Fatalf("[CLIENT] erro ao chamar servidor: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Fatalf("[CLIENT] servidor retornou status %d", resp.StatusCode)
    }

    var out cotacaoOut
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        log.Fatalf("[CLIENT] erro decodificando resposta: %v", err)
    }

    if out.Bid == "" {
        log.Fatal(errors.New("[CLIENT] bid vazio na resposta"))
    }

    // Salva no arquivo (ASCII-only label)
    content := fmt.Sprintf("Dolar: %s\n", out.Bid)
    if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
        log.Fatalf("[CLIENT] erro escrevendo arquivo: %v", err)
    }

    log.Printf("Cotacao salva em %s: %s", outFile, out.Bid)
}

