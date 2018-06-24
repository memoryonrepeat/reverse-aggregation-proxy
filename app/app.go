package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	// "os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_TOP         = 10
	DEFAULT_TOP     = 5
	DEFAULT_SKIP    = 0
	BASE_URL        = "https://s3-eu-west-1.amazonaws.com/test-golang-recipes/"
	REQUEST_TIMEOUT = 2
	CLIENT_TIMEOUT  = 10
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

	ids := make([]string, top)
	for i := range ids {
		ids[i] = strconv.Itoa(i + skip + 1)
	}
	recipes := fetchRecipeList(&ids, client)
	return recipes
}

func AggregatedRecipeHandler(req *http.Request, client *http.Client) []Recipe {
	ids := strings.Split(req.URL.Query()["ids"][0], ",")
	recipes := fetchRecipeList(&ids, client)
	sort.Sort(ByPrepTime(recipes))
	return recipes
}

// Should redirect to AllRecipeHandler / AggregatedRecipeHandler
func ReverseAggregatorProxy(w http.ResponseWriter, req *http.Request) {

	// Specify timeout to avoid apps to hang unexpecedly since there is no timeout by default
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	httpClient := &http.Client{
		Timeout: time.Second * CLIENT_TIMEOUT,
	}
	var recipe []Recipe
	if len(req.URL.Query()["ids"]) > 0 {
		recipe = AggregatedRecipeHandler(req, httpClient)
	} else {
		recipe = AllRecipeHandler(req, httpClient)
	}
	obj, err := json.Marshal(recipe)
	checkError(err)
	io.WriteString(w, string(obj))
}

func fetchSingleRecipe(url string, client *http.Client) Recipe {
	req, err := http.NewRequest("GET", url, nil)
	checkError(err)
	req.Header.Set("User-Agent", "hellofresh")
	resp, err := client.Do(req)
	checkError(err)
	fmt.Println("URL:", url, "|| Status Code:", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Recipe{}
	}
	body, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	var recipe Recipe
	err = json.Unmarshal(body, &recipe)
	checkError(err)
	return recipe
}

func fetchRecipeList(ids *[]string, client *http.Client) []Recipe {
	var recipes []Recipe

	c := make(chan Recipe)

	timeout := time.After(REQUEST_TIMEOUT * time.Second)

	for _, id := range *ids {
		go (func(id string) {
			c <- fetchSingleRecipe(BASE_URL+id, client)
		})(id)
	}
	for _, id := range *ids {
		select {
		case recipe := <-c:
			if recipe.Id != "" {
				recipes = append(recipes, recipe)
			}
		case <-timeout:
			fmt.Println("A request timed out", id)
		}
	}
	return recipes
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		// os.Exit(1)
	}
}
