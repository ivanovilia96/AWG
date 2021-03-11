package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func getData(reciever chan float64) {

	for {
		sleepingTime := (rand.Float64() * 2.5) + 0.5
		sleepingTimeSeconds := int(sleepingTime)
		// отнимаю от всего времени кол-во секунд -> получется всегда 0.х (если мы это просто *10 что бы было 1 целое, то будет не точно),
		// вывожу из под нуля -> что бы потом на это время добавить задержки
		sleepingTimeMicroseconds, err := strconv.Atoi(fmt.Sprintf("%f", sleepingTime-float64(sleepingTimeSeconds))[2:])
		if err != nil {
			sleepingTimeMicroseconds = 0
		}
		// привожу к нужному типу данных
		sleepSeconds := time.Duration(sleepingTime) * time.Second
		sleepMicroseconds := time.Duration(sleepingTimeMicroseconds) * time.Microsecond
		time.Sleep(sleepSeconds + sleepMicroseconds)
		data := (rand.Float64() * 9) + 1
		reciever <- data
	}

}

func main() {
	var (
		recieverChan             = make(chan float64)
		avgValue, sumOfAllFloats float64
		countOfAllRecievedDatas  int
	)

	go getData(recieverChan)

	// раз в 5 секунд будет выводиться среднее значение
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Printf("your AVG value is: %v \n", avgValue)
		}
	}()

	// это что бы при ctrl c вы получали текущее среднее значение
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Printf("your AVG value is: %v , i`m ending \n", avgValue)
			os.Exit(1)
		}
	}()

	for {
		select {
		case val := <-recieverChan:
			sumOfAllFloats += val
			countOfAllRecievedDatas++
			avgValue = sumOfAllFloats / float64(countOfAllRecievedDatas)
			fmt.Println(val, " - recieved value")
		}
	}
}
