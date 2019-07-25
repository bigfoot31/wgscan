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

var skipDistance = 100

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
	// load configuration
	configuration := loadConfig()

	goroutinesController := make(chan int, nthreads)

	// for each chromosome in configuration
	// create a goroutine
	// number of goroutines controlled by
	// channel `goroutinesController`
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

	noCalculations := int(math.Floor(float64(len(inputArray)-4000)/float64(skipDistance))) + 1

	avg1 := calcWindowMean(inputArray, windowStart, windowWidth, noCalculations)
	avg2 := calcWindowMean(inputArray, upstreamStart, upstreamWidth, noCalculations)
	avg3 := calcWindowMean(inputArray, downstreamStart, downstreamWidth, noCalculations)

	count := 0
	results := make([]float64, len(avg1))
	for index := range avg1 {
		results[index] = avg1[index] / ((avg2[index] + avg3[index]) / 2)

		if results[index] > 0 {
			count++
		}
	}

	writeFile(val.Output, results)

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

func writeFile(fileName string, data []float64) {
	file, _ := os.Create(fileName)
	defer file.Close()

	dataFloat32 := make([]float32, len(data))
	for index := range data {
		dataFloat32[index] = float32(data[index])
	}

	file.WriteString(strings.Trim(strings.Replace(fmt.Sprintf("%.2f", dataFloat32), " ", ",", -1), "[]"))
}

func calcWindowMean(inputArray []float64, startIndex int, windowWidth int, noCalculations int) []float64 {
	// where will the loop end
	//
	// for 4000, end = 0
	// for 4001, end = 0
	// for 4000+skipDistance, end = 4000
	endIndex := (noCalculations-1)*skipDistance + startIndex

	// how many values will be outputted
	//
	// for 4000, no = 1
	// for 4001, no = 1
	// for 4000+skipDistance, no = 2
	avg := make([]float64, noCalculations)

	// loop from start to end, with skipDistance jump
	index := 0
	for i := startIndex; i <= endIndex; i += skipDistance {
		sum := 0.0
		// for each skipDistance instance calc avg over windowWidth
		for j := i; j < i+windowWidth && j < len(inputArray); j++ {
			sum = sum + inputArray[j]
		}

		avg[index] = sum / float64(windowWidth)
		index++
	}

	return avg
}
