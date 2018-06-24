package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_TOP      = 10
	DEFAULT_TOP  = 5
	DEFAULT_SKIP = 0
	BASE_URL     = "https://s3-eu-west-1.amazonaws.com/test-golang-recipes/"
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

type ByPrepTime []Recipe

func main() {
	http.HandleFunc("/recipes", ReverseAggregatorProxy)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (p ByPrepTime) Len() int {
	return len(p)
}

func (p ByPrepTime) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p ByPrepTime) Less(i, j int) bool {
	return p[i].PrepTime < p[j].PrepTime
}

func AllRecipeHandler(req *http.Request, client *http.Client) []Recipe {
	top := DEFAULT_TOP
	skip := DEFAULT_SKIP
	if len(req.URL.Query()["top"]) > 0 {
		top, _ = strconv.Atoi(req.URL.Query()["top"][0])
		if top > MAX_TOP {
			top = MAX_TOP
		}
	}
	if len(req.URL.Query()["skip"]) > 0 {
		skip, _ = strconv.Atoi(req.URL.Query()["skip"][0])
	}

	var recipes []Recipe
	c := make(chan Recipe)
	timeout := time.After(2 * time.Second)

	for i := skip; i < (top + skip); i++ {
		go (func(i int) {
			c <- fetchSingleRecipe(BASE_URL+strconv.Itoa(i+1), client)
		})(i)
	}
	for i := skip; i < (top + skip); i++ {
		select {
		case recipe := <-c:
			if recipe.Id != "" {
				recipes = append(recipes, recipe)
			}
		case <-timeout:
			fmt.Println("timeout")
			// return recipes
		}
		/*recipe := <-c
		recipes = append(recipes, recipe)*/
	}
	return recipes
}

func AggregatedRecipeHandler(req *http.Request, client *http.Client) []Recipe {
	ids := strings.Split(req.URL.Query()["ids"][0], ",")
	var recipes []Recipe
	c := make(chan Recipe)

	timeout := time.After(2 * time.Second)

	for _, id := range ids {
		go (func(id string) {
			fmt.Println(id)
			c <- fetchSingleRecipe(BASE_URL+id, client)
		})(id)
	}
	for _, id := range ids {
		select {
		case recipe := <-c:
			if recipe.Id != "" {
				recipes = append(recipes, recipe)
			}
		case <-timeout:
			fmt.Println("timeout", id)
			// return recipes
		}
		/*recipe := <-c
		recipes = append(recipes, recipe)*/
	}
	sort.Sort(ByPrepTime(recipes))
	return recipes
}

// Should redirect to AllRecipeHandler / AggregatedRecipeHandler
func ReverseAggregatorProxy(w http.ResponseWriter, req *http.Request) {

	// Specify timeout to avoid apps to hang unexpecedly since there is no timeout by default
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	httpClient := &http.Client{
		Timeout: time.Second * 2,
	}
	var r []Recipe
	if len(req.URL.Query()["ids"]) > 0 {
		r = AggregatedRecipeHandler(req, httpClient)
	} else {
		r = AllRecipeHandler(req, httpClient)
	}
	j, e := json.Marshal(r)
	check(e)
	io.WriteString(w, string(j))
}

func fetchSingleRecipe(url string, client *http.Client) Recipe {
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	check(err)
	req.Header.Set("User-Agent", "hellofresh")
	resp, err := client.Do(req)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Recipe{}
	}
	fmt.Println("HTTP Response Status:", resp.StatusCode, "||", http.StatusText(resp.StatusCode))
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
		fmt.Println("bug", err)
		os.Exit(1)
	}
}
