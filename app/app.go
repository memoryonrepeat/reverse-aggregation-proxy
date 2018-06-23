package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Recipe struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
	Difficulty  int    `json:"difficulty"`
	PrepTime    string `json:"prepTime"`
	ImageLink   string `json:"imageLink"`
	Ingredients []struct {
		Name      string `json:"name"`
		ImageLink string `json:"imageLink"`
	} `json:"ingredients"`
}

func main() {
	http.HandleFunc("/", ReverseAggregatorProxy)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ReverseAggregatorProxy(w http.ResponseWriter, req *http.Request) {
	r := fetch()
	j, e := json.Marshal(r)
	check(e)
	io.WriteString(w, string(j))
}

func fetch() []Recipe {
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
			c <- fetchSingleRecipe("https://s3-eu-west-1.amazonaws.com/test-golang-recipes/"+strconv.Itoa(i), httpClient)
		})(i)
	}
	for i := 0; i < 4; i++ {
		select {
		case recipe := <-c:
			recipes = append(recipes, recipe)
		case <-timeout:
			fmt.Println("timeout")
			// return recipes
		}
		/*recipe := <-c
		recipes = append(recipes, recipe)*/
	}
	return recipes
}

func fetchSingleRecipe(url string, client *http.Client) Recipe {
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
