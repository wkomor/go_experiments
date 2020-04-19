package main

import (
	"sort"
	"strconv"
	"fmt"
	"time"
	"sync"
	"strings"
	// "runtime"
)

type defaultSort []string

func (a defaultSort) Len() int           { return len(a) }
func (a defaultSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a defaultSort) Less(i, j int) bool { return a[i] < a[j] }


// SingleHash func
//func SingleHash(data string, sh chan<- string) {
func SingleHash(data string, sh chan<- string) {
	mu := &sync.Mutex{}
	mu.Lock()
	md := DataSignerMd5(data)
	l :=  DataSignerCrc32(data)
	r :=  DataSignerCrc32(md)
	res := l + "~" + r
	mu.Unlock()
	sh <- res
}

// MultiHash func
func MultiHash(data string, results chan string) {	
	var wg sync.WaitGroup 
	wg.Add(6)      
	multiHash := ""
	hashed := ""
	ch := make(chan string, 100)
	sch := make(chan string)
	go SingleHash(data, sch)
	hashed = <- sch
	iterHash := func (item int, hashed string, ch chan<- string)  {
		defer wg.Done()
		ch <-  DataSignerCrc32(strconv.Itoa(item) + hashed)
	}
	for i := 0; i < 6; i++ {
		go iterHash(i, hashed, ch)
		multiHash += <- ch
	}
	results <- multiHash
	wg.Wait()
	
}

// CombineResults func
func CombineResults(data []string) string {
	go sort.Sort(defaultSort(data))
	r := strings.Join(data, "_")
	return r
}


// ExecutePipeline func
func ExecutePipeline(data []int) {
	var results []string
	ch := make(chan string, 100)
	start := time.Now()
	for _, i := range data {
		go MultiHash(strconv.Itoa(i), ch)
		results = append(results, <-ch)
		// runtime.Gosched()
	}
	h := CombineResults(results)
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(elapsed)
	fmt.Println(h)
}

func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	//inputData := []int{0, 1}
	ExecutePipeline(inputData)
}
