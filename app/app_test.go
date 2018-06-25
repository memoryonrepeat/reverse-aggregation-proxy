package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func TestReverseAggregatorProxyHandler1(t *testing.T) {
	req, err := http.NewRequest("GET", "/recipes", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ReverseAggregatorProxyHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestAllRecipeHandler1(t *testing.T) {
	req, err := http.NewRequest("GET", "/recipes?top=10&skip=-5", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ReverseAggregatorProxyHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestAggregatedRecipeHandler2(t *testing.T) {
	req, err := http.NewRequest("GET", "/recipes?ids=1,2,3a,-4", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ReverseAggregatorProxyHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestFilter(t *testing.T) {
	type args struct {
		input []string
		f     func(string) bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"Filter out non-positive non-int strings", args{
				[]string{"a", "a1", "1", "-1", "2", "1a"},
				func(s string) bool {
					val, _ := strconv.Atoi(s)
					return val > 0
				},
			}, []string{"1", "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.args.input, tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByPrepTime_Len(t *testing.T) {
	tests := []struct {
		name string
		p    ByPrepTime
		want int
	}{
		{"Normal array", []Recipe{
			{}, {}, {},
		}, 3},
		{"Empty array", []Recipe{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Len(); got != tt.want {
				t.Errorf("ByPrepTime.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByPrepTime_Less(t *testing.T) {

	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		p    ByPrepTime
		args args
		want bool
	}{
		{"30M < 50M", []Recipe{
			{PrepTime: "PT30M"},
			{PrepTime: "PT50M"},
		}, args{0, 1}, true},
		{"45M < 1H", []Recipe{
			{PrepTime: "PT45M"},
			{PrepTime: "PT1H"},
		}, args{0, 1}, true},
		{"61M > 1H", []Recipe{
			{PrepTime: "PT61M"},
			{PrepTime: "PT1H"},
		}, args{0, 1}, false},
		{"60M == 1H", []Recipe{
			{PrepTime: "PT60M"},
			{PrepTime: "PT1"},
		}, args{0, 1}, false},
		{"2H > 1H", []Recipe{
			{PrepTime: "PT2H"},
			{PrepTime: "PT1H"},
		}, args{0, 1}, false},
		{"1H30M < 1H40M", []Recipe{
			{PrepTime: "PT1H30M"},
			{PrepTime: "PT1H40M"},
		}, args{0, 1}, true},
		{"1H40M == 1H40M", []Recipe{
			{PrepTime: "PT1H30M"},
			{PrepTime: "PT1H30M"},
		}, args{0, 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("ByPrepTime.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}
