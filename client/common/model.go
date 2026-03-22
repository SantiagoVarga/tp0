package common

import (
	"fmt"
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
