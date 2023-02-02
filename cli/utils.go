package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

func IsEmptyString(str any) bool {
	if realStr, ok := str.(string); ok {
		return len(strings.TrimSpace(realStr)) == 0
	}
	return true
}

func BigIntToInt(bigInt *big.Int) int {
	integer, _ := strconv.Atoi(bigInt.String())
	return integer
}

func fetchImage(url string) (image.Image, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()
	return image.Decode(resp.Body)
}

func StatusToText(completed bool) string {
	if completed {
		return "Completed"
	}
	return "Ongoing"
}
