package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

type BetInfo struct {
	Agency   string
	Name     string
	Surname  string
	DNI      string
	Birthday string
	Number   string
}

type ServerResponse struct {
	Type    string
	Status  string
	Message string
}

type BatchConfig struct {
	MaxAmount int
}

func LoadBetInfo(clientID string) (BetInfo, error) {
	var betInfo BetInfo

	betInfo.Agency = clientID
	betInfo.Name = os.Getenv("NAME")
	betInfo.Surname = os.Getenv("SURNAME")
	betInfo.DNI = os.Getenv("DNI")
	betInfo.Birthday = os.Getenv("BIRTHDAY")
	number := os.Getenv("NUMBER")
	if number == "" {
		return betInfo, fmt.Errorf("NUMBER environment variable is not set")
	}
	betInfo.Number = number

	if betInfo.Name == "" || betInfo.Surname == "" || betInfo.DNI == "" || betInfo.Birthday == "" {
		return betInfo, fmt.Errorf("One or more required environment variables (NAME, SURNAME, DNI, BIRTHDAY) are not set")
	}

	return betInfo, nil
}

func LoadBetsFromFile(clientID string, filePath string) ([]BetInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("action: read_bet_file | result: fail |  error: %v", err)
		return nil, err
	}
	// Ensure the file is closed after reading
	defer file.Close()

	reader := csv.NewReader(file)
	var bets []BetInfo

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		bet := BetInfo{
			Agency:   clientID,
			Name:     record[0],
			Surname:  record[1],
			DNI:      record[2],
			Birthday: record[3],
			Number:   record[4],
		}
		bets = append(bets, bet)
	}

	return bets, nil
}
