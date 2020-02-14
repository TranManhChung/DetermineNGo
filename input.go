package main

import "time"

func doWork(done <-chan interface{}, queue chan int, id int) int {
	var numOfJob int
	for {
		select {
		case <-queue:
			//-------------------
			time.Sleep(10 * time.Millisecond)
			//-------------------
			numOfJob++
		case <-done:
			// fmt.Println("id : ", id)
			return numOfJob
		}
	}
}
