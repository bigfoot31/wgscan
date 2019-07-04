package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

var nthreads = runtime.NumCPU()

var configFile = "config.json"

var windowStart = 1950
var windowWidth = 101

var upstreamStart = 0
var upstreamWidth = 1000

var downstreamStart = 3000
var downstreamWidth = 1000

type interfaceConfiguration struct {
	File   string
	Output string
	Length int
}

func main() {
	configuration := loadConfig()

	goroutinesController := make(chan int, nthreads)

	for _, val := range configuration {
		goroutinesController <- 1
		wg.Add(1)
		go driver(val, goroutinesController)
	}

	wg.Wait()
}

func loadConfig() []interfaceConfiguration {
	file, _ := os.Open(configFile)
	defer file.Close()

	configuration := []interfaceConfiguration{}
	err := json.NewDecoder(file).Decode(&configuration)

	if err != nil {
		fmt.Println("config load error:", err)
	}

	return configuration
}

func driver(val interfaceConfiguration, goroutinesController chan int) {
	inputArray := readFile(val.File, val.Length)

	avg1 := calcWindowMean(inputArray, windowStart, windowWidth)
	avg2 := calcWindowMean(inputArray, upstreamStart, upstreamWidth)
	avg3 := calcWindowMean(inputArray, downstreamStart, downstreamWidth)

	count := 0
	results := make([]float64, len(avg1))
	for index := range avg1 {
		results[index] = avg1[index] / ((avg2[index] + avg3[index]) / 2)

		if results[index] > 0 {
			count++
		}
	}

	writer(val.Output, results)

	<-goroutinesController
	wg.Done()
}

func readFile(fileName string, size int) []float64 {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("io error:", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	inputArray := make([]float64, size)
	line, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println(err)
	}

	for index, val := range strings.Split(string(line), ",") {
		tmp, _ := strconv.ParseFloat(val, 32)
		inputArray[index] = float64(tmp)
	}

	return inputArray
}

func writer(fileName string, data []float64) {
	file, _ := os.Create(fileName)
	defer file.Close()

	dataFloat32 := make([]float32, len(data))
	for index := range data {
		dataFloat32[index] = float32(data[index])
	}

	file.WriteString(strings.Trim(strings.Replace(fmt.Sprintf("%.2f", dataFloat32), " ", ",", -1), "[]"))
}

func calcWindowMean(inputArray []float64, start int, window int) []float64 {
	end := len(inputArray) - (4001 - start - windowWidth)
	avg := make([]float64, int(math.Ceil((float64(end-start-windowWidth+1)/float64(windowWidth))))+1)

	index := 0
	for i := start; i <= end; i += windowWidth {
		sum := 0.0
		for j := i; j < i+window && j < len(inputArray); j++ {
			sum = sum + inputArray[j]
		}

		avg[index] = sum / float64(window)
		index++
	}

	return avg
}
