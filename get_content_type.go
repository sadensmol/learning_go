package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	ct, err := getContentType("https://ya.ru")

	if err != nil {
		fmt.Println(err)

	} else {
		fmt.Printf("content type: %s", ct)
	}

}

func getContentType(url string) (string, error) {
	fmt.Println(url)
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return "", fmt.Errorf("error!")

	} else {

		rrr, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error 222")

		} else {
			return http.DetectContentType(rrr), nil
		}

	}

}
