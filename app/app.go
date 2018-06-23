package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
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
	// Specify timeout to avoid apps to hang unexpecedly since there is no timeout by default
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	httpClient := &http.Client{
		Timeout: time.Second * 2,
	}
	var recipes []Recipe
	c := make(chan Recipe)
	timeout := time.After(2 * time.Second)
	for i := 1; i < 5; i++ {
		go (func(i int) {
			c <- fetch("https://s3-eu-west-1.amazonaws.com/test-golang-recipes/"+strconv.Itoa(i), httpClient)
		})(i)
	}
	for i := 0; i < 4; i++ {
		select {
		case recipe := <-c:
			recipes = append(recipes, recipe)
		case <-timeout:
			fmt.Println("timeout")
			return
		}
		/*recipe := <-c
		recipes = append(recipes, recipe)*/
	}
	fmt.Println(recipes)
}

func fetch(url string, client *http.Client) Recipe {
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	check(err)
	req.Header.Set("User-Agent", "hellofresh")
	resp, err := client.Do(req)
	check(err)
	body, err := ioutil.ReadAll(resp.Body)
	check(err)
	var recipe Recipe
	err = json.Unmarshal(body, &recipe)
	check(err)
	return recipe
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
