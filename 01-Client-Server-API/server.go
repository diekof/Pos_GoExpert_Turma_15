// server.go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const (
	apiURL             = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	httpAddr           = ":8080"
	dbFile             = "cotacoes.db"
	apiTimeout         = 200 * time.Millisecond
	dbInsertTimeout    = 10 * time.Millisecond
	serverReadTimeout  = 5 * time.Second
	serverWriteTimeout = 5 * time.Second
)

type awesomeResp struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type cotacaoOut struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("erro abrindo SQLite: %v", err)
	}
	defer db.Close()

	if err := createSchema(db); err != nil {
		log.Fatalf("erro criando schema: %v", err)
	}

    mux := http.NewServeMux()
    mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
        handleCotacaoV2(w, r, db)
    })

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      logging(mux),
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
	}

	log.Printf("Servidor ouvindo em %s ...", httpAddr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
		log.Fatalf("erro no servidor: %v", err)
	}
}

func createSchema(db *sql.DB) error {
	const ddl = `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		obtida_em TIMESTAMP NOT NULL
	);`
	_, err := db.Exec(ddl)
	return err
}

// handleCotacaoV2: versão ASCII-only para evitar problemas de encoding
func handleCotacaoV2(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    ctxAPI, cancelAPI := context.WithTimeout(r.Context(), apiTimeout)
    defer cancelAPI()

    bid, err := fetchBid(ctxAPI)
    if err != nil {
        log.Printf("[API] erro: %v", err)
        http.Error(w, "falha ao obter cotacao", http.StatusGatewayTimeout)
        return
    }

    ctxDB, cancelDB := context.WithTimeout(r.Context(), dbInsertTimeout)
    defer cancelDB()
    if err := saveBid(ctxDB, db, bid); err != nil {
        log.Printf("[DB] erro ao persistir: %v", err)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(cotacaoOut{Bid: bid})
}

func handleCotacao(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), apiTimeout)
	defer cancelAPI()

	bid, err := fetchBid(ctxAPI)
	if err != nil {
		log.Printf("[API] erro: %v", err)
		http.Error(w, "falha ao obter cotação", http.StatusGatewayTimeout)
		return
	}

	ctxDB, cancelDB := context.WithTimeout(r.Context(), dbInsertTimeout)
	defer cancelDB()
	if err := saveBid(ctxDB, db, bid); err != nil {
		log.Printf("[DB] erro ao persistir: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacaoOut{Bid: bid})
}

func fetchBid(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("resposta não OK da AwesomeAPI")
	}

	var a awesomeResp
	if err := json.NewDecoder(resp.Body).Decode(&a); err != nil {
		return "", err
	}

	if a.USDBRL.Bid == "" {
		return "", errors.New("campo bid vazio na resposta")
	}

	return a.USDBRL.Bid, nil
}

func saveBid(ctx context.Context, db *sql.DB, bid string) error {
	const ins = `INSERT INTO cotacoes (bid, obtida_em) VALUES (?, ?)`
	_, err := db.ExecContext(ctx, ins, bid, time.Now().UTC())
	return err
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		dur := time.Since(start)
		log.Printf("%s %s - %v", r.Method, r.URL.Path, dur)
	})
}
