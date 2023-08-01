package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Exchange struct {
	ID      uint   `gorm:"primaryKey"`
	Cotacao string `gorm:"column:cotacao"`
}

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
	println("Server running on port 8080")
	http.HandleFunc("/cotacao", getExangeHandler)
	http.ListenAndServe(":8080", nil)
}

func getExangeHandler(w http.ResponseWriter, r *http.Request) {
	s, error := GetDollarToRealExange()
	if error != nil {
		panic(error)
	}

	saveOnDatabase(&s.USDBRL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(*s)
}

func GetDollarToRealExange() (*ExchangeAPI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, error := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if error != nil {
		log.Printf("Time out error.\n[ERRO] - %v", error)
		return nil, error
	}

	if ctx.Err() != nil {
		log.Printf("Context error: %v", ctx.Err())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error on response.\n[ERRO] - %v", err)
		return nil, error
	}
	defer res.Body.Close()

	body, error := ioutil.ReadAll(res.Body)
	if error != nil {
		return nil, error
	}

	var e ExchangeAPI
	error = json.Unmarshal(body, &e)
	if error != nil {
		return nil, error
	}

	return &e, nil
}

func saveOnDatabase(e *ExchangeResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	dsn := "root:root@tcp(localhost:3306)/goexpert"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return err
	}

	db.WithContext(ctx)

	db.AutoMigrate(&Exchange{})

	if e.Bid == "" {
		return errors.New("bid value is empty")
	}

	cotacao, err := strconv.ParseFloat(e.Bid, 64)
	if err != nil {
		return err
	}

	exchange := &Exchange{
		Cotacao: fmt.Sprintf("%.2f", cotacao),
	}

	err = db.Create(&exchange).Error
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		log.Printf("Context error: %v", ctx.Err())
	}

	if err != nil {
		return err
	}

	return nil
}
