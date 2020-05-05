package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	// "runtime"
)

type defaultSort []string

func (a defaultSort) Len() int           { return len(a) }
func (a defaultSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a defaultSort) Less(i, j int) bool { return a[i] < a[j] }

func calcSingleHash(data string, ch chan string) {
	ch <- DataSignerCrc32(data)
}

// SingleHash func
//func SingleHash(data string, sh chan<- string) {
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for rawData := range in {
		data := strconv.Itoa(rawData.(int))
		dataMd5 := DataSignerMd5(data)
		wg.Add(1)
		go func() {
			leftChan := make(chan string)
			rightChan := make(chan string)
			go calcSingleHash(data, leftChan)
			go calcSingleHash(dataMd5, rightChan)
			leftCrc32 := <-leftChan
			rightCrc32 := <-rightChan
			out <- leftCrc32 + "~" + rightCrc32
			wg.Done()
		}()
	}
}

func calcMultiHash(item int, multiHash []string, wg *sync.WaitGroup, data string) {
	multiHash[item] = DataSignerCrc32(strconv.Itoa(item) + data)
	wg.Done()
}

// MultiHash func
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for rawData := range in {
		wg.Add(1)
		go func(data string) {
			innerWG := &sync.WaitGroup{}
			multiHash := make([]string, 6)

			for item := 0; item < 6; item++ {
				innerWG.Add(1)
				go calcMultiHash(item, multiHash, innerWG, data)
			}
			innerWG.Wait()
			out <- strings.Join(multiHash, "")
			wg.Done()
		}(rawData.(string))
	}
}

// CombineResults func
func CombineResults(in, out chan interface{}) {
	var toSort []string

	for data := range in {
		toSort = append(toSort, data.(string))
	}
	sort.Strings(toSort)

	out <- strings.Join(toSort, "_")
}

// ExecutePipeline func
func ExecutePipeline(hashJobs ...job) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ch := make(chan interface{})
	start := time.Now()
	for _, singleJob := range hashJobs {
		wg.Add(1)
		out := make(chan interface{})
		go func(jobFunc job, in, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			jobFunc(in, out)
		}(singleJob, ch, out, wg)
		ch = out
	}

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(elapsed)

}

func main() {}

// func main() {
// 	inputData := []int{0, 1, 1, 2, 3, 5, 8}
// 	testResult := ""
// 	hashSignJobs := []job{
// 		job(func(in, out chan interface{}) {
// 			for _, fibNum := range inputData {
// 				out <- fibNum
// 			}
// 		}),
// 		job(SingleHash),
// 		job(MultiHash),
// 		job(CombineResults),
// 		job(func(in, out chan interface{}) {
// 			dataRaw := <-in
// 			data, ok := dataRaw.(string)
// 			if !ok {
// 				fmt.Println("cant convert result data to string")
// 			}
// 			testResult = data
// 		}),
// 	}

// 	ExecutePipeline(hashSignJobs...)

// }
