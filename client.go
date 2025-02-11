package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	serverURL     = "http://localhost:8080/cotacao"
	clientTimeout = 300 * time.Millisecond
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		log.Fatal("Erro ao criar requisição:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao fazer requisição:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Erro ao ler resposta:", err)
	}

	log.Println("Resposta do servidor:", string(body))

	var cotacao Cotacao
	if err := json.Unmarshal(body, &cotacao); err != nil {
		log.Fatal("Erro ao decodificar resposta:", err)
	}

	fmt.Println("Cotação do dólar:", cotacao.Bid)

	if err := salvarCotacaoArquivo(cotacao.Bid); err != nil {
		log.Fatal("Erro ao salvar cotação no arquivo:", err)
	}
}

func salvarCotacaoArquivo(bid string) error {
	content := fmt.Sprintf("Dólar: %s\n", bid)
	return ioutil.WriteFile("cotacao.txt", []byte(content), 0644)
}
