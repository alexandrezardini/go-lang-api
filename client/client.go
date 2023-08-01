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

type ExchangeResponse struct {
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
}

type ExchangeAPI struct {
	USDBRL ExchangeResponse `json:"USDBRL"`
}

func main() {
	cotacao, err := GetExchageValue()
	if err != nil {
		log.Println(err)
	}
	saveBidToFile(cotacao)
}

func GetExchageValue() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, error := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if error != nil {
		log.Printf("Time out error.\n[ERRO] - %v", error)
		return "", error
	}

	if ctx.Err() != nil {
		log.Printf("Context error: %v", ctx.Err())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error on response.\n[ERRO] - %v", err)
		return "", error
	}
	defer res.Body.Close()

	body, error := ioutil.ReadAll(res.Body)
	if error != nil {
		return "", error
	}

	var e ExchangeAPI
	error = json.Unmarshal(body, &e)
	if error != nil {
		return "", error
	}

	var bid = &e.USDBRL.Bid

	return *bid, nil
}

func saveBidToFile(bid string) error {
	str := fmt.Sprintf("Dólar: %s", bid)

	err := ioutil.WriteFile("cotacao.txt", []byte(str), 0644)
	if err != nil {
		return err
	}

	return nil
}
