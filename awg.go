package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

func getData(reciever chan float64) {
	for {

		var sleepingTimeSeconds int
		sleepingTimeMiliseconds := rand.Intn(1001)
		// если sleepingTimeMilisec выпало как 0, то 3 секунды может быть, иначе максимум 2 и х mili секунд
		if sleepingTimeMiliseconds == 0 {
			sleepingTimeSeconds = rand.Intn(4)
		} else {
			sleepingTimeSeconds = rand.Intn(3)
		}

		// привожу к нужному типу данных
		sleepSeconds := time.Duration(sleepingTimeSeconds) * time.Second
		sleepMicroseconds := time.Duration(sleepingTimeMiliseconds) * time.Millisecond
		time.Sleep(sleepSeconds + sleepMicroseconds)
		data := (rand.Float64() * 9) + 1
		reciever <- data
	}
}

func main() {
	var (
		recieverChan            = make(chan float64)
		avgValue                float64
		countOfAllRecievedDatas int
	)

	go getData(recieverChan)

	// это что бы при ctrl c вы получали текущее среднее значение
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)

	// раз в 5 секунд будет выводиться среднее значение
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case val := <-recieverChan:
			countOfAllRecievedDatas++
			fmt.Println(val, " - recieved value")
			avgValue += (val - avgValue) / float64(countOfAllRecievedDatas)
		case <-ticker.C:
			fmt.Println("Current awg: ", avgValue)
		case <-interruptSignal:
			fmt.Printf("your AVG value is: %v , i`m ending \n", avgValue)
			os.Exit(1)
		}
	}
}
