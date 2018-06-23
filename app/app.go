package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Recipe struct {
	Id          string
	Name        string
	Headline    string
	Description string
	Difficulty  int
	PrepTime    string
	ImageLink   string
	Ingredients []struct {
		Name      string
		ImageLink string
	}
}

func main() {
	resp, err := http.Get("https://s3-eu-west-1.amazonaws.com/test-golang-recipes/1")
	check(err)
	body, err := ioutil.ReadAll(resp.Body)
	check(err)
	var recipe Recipe
	err = json.Unmarshal(body, &recipe)
	check(err)
	fmt.Println(recipe)
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
