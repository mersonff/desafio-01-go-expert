package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"io/ioutil"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL       = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	httpTimeout  = 200 * time.Millisecond
	dbTimeout    = 10 * time.Millisecond
	serverPort   = ":8080"
	databaseFile = "cotacoes.db"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatal("Erro ao abrir banco de dados:", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		bid TEXT, 
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal("Erro ao criar tabela:", err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		cotacao, err := buscarCotacao()
		if err != nil {
			http.Error(w, "Erro ao obter cotação", http.StatusInternalServerError)
			log.Println("Erro ao obter cotação:", err)
			return
		}

		err = salvarCotacao(db, cotacao.Bid)
		if err != nil {
			log.Println("Erro ao salvar cotação no banco:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cotacao)
	})

	log.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(serverPort, nil))
}

func buscarCotacao() (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao requisitar API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na resposta da API: status %d - %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da API: %v", err)
	}

	log.Println("Resposta da API:", string(body))

	var result map[string]Cotacao
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	cotacao, ok := result["USDBRL"]
	if !ok {
		return nil, fmt.Errorf("resposta inesperada da API, chave 'USDBRL' não encontrada")
	}

	return &cotacao, nil
}

func salvarCotacao(db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
	if err != nil {
		return fmt.Errorf("erro ao inserir no banco: %v", err)
	}

	return nil
}
