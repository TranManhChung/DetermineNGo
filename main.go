package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type startGoroutineFn func(done <-chan interface{}, queue chan int, id int) int

func setupConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func simulationJobQueue() chan int {
	queue := make(chan int, 1000)
	go func() {
		for {
			queue <- 0
		}
	}()
	return queue
}

// queue : queue jobs
// numGoIncr : number of goroutines increase per roud
func coreProcessing(timeout time.Duration, startGoroutine startGoroutineFn, queue chan int, numGoIncr int) {
	queueCompile := make(chan int, viper.GetInt("goroutine.max_size_test"))
	numOfGos := 1
	printAmount := make(chan interface{})
	compileDone := make(chan interface{})
	go func() {
		var wardDone chan interface{}
		wardDone = make(chan interface{})
		startWard := func(queue chan int, wardDone chan interface{}, id int) {
			queueCompile <- startGoroutine(wardDone, queue, id)
		}
		//--------------------------------------------- Print the result
		go func(queueCompile chan int, printAmount chan interface{}) {
			for {
				numOfJobs := 0
				<-printAmount
				idx := 1
				for numOfJob := range queueCompile {
					numOfJobs += numOfJob
					idx++
					if idx > numOfGos {
						break
					}
				}
				fmt.Printf("Total jobs: %d / Total gorountines: %d / Timeout : %s", numOfJobs, numOfGos, viper.GetDuration("goroutine.timeout_each_round"))
				compileDone <- struct{}{}
			}
		}(queueCompile, printAmount)
		//---------------------------------------------
		//----start a sample round
		go startWard(queue, wardDone, -1)
		//
		round := 0
	monitorLoop:
		for {
			for {
				round++
				<-time.After(timeout)
				close(wardDone)
				printAmount <- struct{}{}
				wardDone = make(chan interface{})
				<-compileDone
				numOfGos += numGoIncr
				fmt.Println("\nStart new round!!!")
				for i := 0; i < numOfGos; i++ {
					go startWard(queue, wardDone, round)
				}
				continue monitorLoop
			}
		}
	}()
}

func main() {

	setupConfig()
	queue := simulationJobQueue()

	coreProcessing(viper.GetDuration("goroutine.timeout_each_round"), doWork, queue, viper.GetInt("goroutine.num_incr"))
	time.Sleep(1 * time.Hour)
}
