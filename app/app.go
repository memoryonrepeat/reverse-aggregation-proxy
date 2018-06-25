package main

import (
	config "./config"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

type ByPrepTime []Recipe

func main() {
	http.HandleFunc("/recipes", ReverseAggregatorProxy)
	log.Println("Server starting at port", config.Port)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Redirects the request to AllRecipeHandler / AggregatedRecipeHandler
func ReverseAggregatorProxy(w http.ResponseWriter, req *http.Request) {
	log.Println("Incoming request", req.URL)
	// Specify timeout to avoid apps to hang unexpecedly since there is no timeout by default
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	httpClient := &http.Client{
		Timeout: time.Second * config.ClientTimeout,
	}
	var recipe []Recipe
	if len(req.URL.Query()["ids"]) > 0 {
		recipe = AggregatedRecipeHandler(req, httpClient)
	} else {
		recipe = AllRecipeHandler(req, httpClient)
	}
	obj, err := json.Marshal(recipe)
	if err != nil {
		log.Println("Request error", req.URL, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Println("Request fulfilled", req.URL)
		w.WriteHeader(http.StatusOK)
	}
	io.WriteString(w, string(obj))
}

// Handle all recipes use case
// If pagination settings (top/skip) is not given, make them the default
// Top parameter is bounded at some max constant to avoid overloading the recipe server
func AllRecipeHandler(req *http.Request, client *http.Client) []Recipe {
	top := config.DefaultTop
	skip := config.DefaultSkip
	if len(req.URL.Query()["top"]) > 0 {
		top, _ = strconv.Atoi(req.URL.Query()["top"][0])
		if top > config.MaxTop {
			top = config.MaxTop
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

// Fetch recipes with given ids
func AggregatedRecipeHandler(req *http.Request, client *http.Client) []Recipe {
	ids := strings.Split(req.URL.Query()["ids"][0], ",")
	recipes := fetchRecipeList(&ids, client)
	sort.Sort(ByPrepTime(recipes))
	return recipes
}

// Fetch all recipes in a given list using goroutines
func fetchRecipeList(ids *[]string, client *http.Client) []Recipe {
	var recipes []Recipe

	c := make(chan Recipe)

	timeout := time.After(config.RequestTimeout * time.Second)

	for _, id := range *ids {
		go (func(id string) {
			c <- fetchSingleRecipe(config.BaseURL+id, client)
		})(id)
	}
	for range *ids {
		select {
		case recipe := <-c:
			if recipe.Id != "" {
				recipes = append(recipes, recipe)
			}
		case <-timeout:
			log.Println("A request timed out")
		}
	}
	return recipes
}

// Fetch a single URL
func fetchSingleRecipe(url string, client *http.Client) Recipe {
	req, err := http.NewRequest("GET", url, nil)
	checkError(err)
	req.Header.Set("User-Agent", "hellofresh")
	resp, err := client.Do(req)
	checkError(err)
	log.Println(url, "|| Status Code:", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Recipe{}
	}
	body, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	var recipe Recipe
	err = json.Unmarshal(body, &recipe)
	checkError(err)
	defer resp.Body.Close()
	return recipe
}

// Implement sort interface
func (p ByPrepTime) Len() int {
	return len(p)
}

// Implement sort interface
func (p ByPrepTime) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Implement sort interface
func (p ByPrepTime) Less(i, j int) bool {
	return p[i].PrepTime < p[j].PrepTime
}

func checkError(err error) {
	if err != nil {
		log.Println(err)
	}
}
