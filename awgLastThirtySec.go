package main

import (
	"container/list"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

// с помощью массива этих структур мы будем знать сколько времени назад пришла структура и что приносила в виде данных
type storeTimeAndData struct {
	data float64
	time time.Duration
}

// возвращает числа от 1 до 10 с интервалом от 0.5 до 3 секунд
func getData(reciever chan storeTimeAndData) {
	for {
		var sleepingTimeSeconds int
		sleepingTimeMiliseconds := rand.Intn(1000)
		// если sleepingTimeMilisec выпало как 0, то 3 секунды может быть, иначе максимум 2 и х mili секунд
		if sleepingTimeMiliseconds == 0 {
			sleepingTimeSeconds = rand.Intn(4)
		} else {
			sleepingTimeSeconds = rand.Intn(3)
		}

		// устанавливает минимальное значение в пол секунды если у нас получается что оно выпало случайно менее этого времени
		if sleepingTimeSeconds == 0 && sleepingTimeMiliseconds < 500 {
			sleepingTimeMiliseconds = 500
		}

		// привожу к нужному типу данных
		sleepSeconds := time.Duration(sleepingTimeSeconds) * time.Second
		sleepMicroseconds := time.Duration(sleepingTimeMiliseconds) * time.Millisecond
		sleepingTime := sleepSeconds + sleepMicroseconds
		time.Sleep(sleepingTime)
		data := (rand.Float64() * 9) + 1
		combinedTimeAndData := storeTimeAndData{
			data, sleepingTime,
		}
		reciever <- combinedTimeAndData
	}
}

func main() {
	var (
		recieverChan = make(chan storeTimeAndData)
		avgValue     float64
		// поле sumOfAllDatas присутствует потому что итерация по связанному списку при получении каждого нового значения потребует большее кол-во операций 
		// чем просто хранение в 1 переменной, так мы сможем
		// оперировать связанным списком наиболее продуктивно  (у всех этих операций О(1) если -> удалять\добавлять\читать данные только в\из начала\конца)
		sumOfAllDatas                  float64
		// тоже самое и с этой переменной : что бы при каждом новом значении не двигаться по списку и не считать
		numberElapsedSeconds           int64
		listOfDataForLastThirtySeconds = list.New()
	)

	go getData(recieverChan)

	// это что бы при ctrl c вы получали текущее среднее значение
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)

	// раз в 5 секунд будет выводиться среднее значение
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case dataAndTime := <-recieverChan:
			parsedDuration, err := time.ParseDuration(dataAndTime.time.String())
			if err != nil {
				panic("there is an error when parse duration which came from chan")
			}

			// подсчет этих секунд ведется что бы иметь возможность оперативно узнать, когда время вообще перевалило за 30 секунд и пора контроллировать стэк
			numberElapsedSeconds += parsedDuration.Milliseconds()
			listOfDataForLastThirtySeconds.PushFront(dataAndTime)
			sumOfAllDatas += dataAndTime.data

			// данный маневр мной сделан для того, что бы , если у нас пришло 2.5 сек, что является 2.5 т мс , а в начале стэка были таймеры на 500мс,
			// то получается что стэк будет продолжать быть переполненным, (кол-во хранимых секунд будет 32 в данном случае, ну и если так дальше пойдет,
			//	 то может храниться и 50 сек и т.д), для этого этот цикл убирает с начала элементы пока вместимость не окажется в приделах 30 т мс.

			for {
				// 30 т милисек это 30 секунд
				if numberElapsedSeconds > 30000 {
					// удаляем первый элемент и оперируем с его данными
					firstElementOfList := listOfDataForLastThirtySeconds.Remove(listOfDataForLastThirtySeconds.Back()).(storeTimeAndData)
					// парсим первый элемент в массиве( который подлежит удалению) что бы отнять от текущего времени его
					parsedDuration, err = time.ParseDuration(firstElementOfList.time.String())
					// отнимаем от суммы  то значение, которое выпадает по времени
					sumOfAllDatas -= firstElementOfList.data
					if err != nil {
						panic("there is an error when parse duration which came from stored data")
					}
					numberElapsedSeconds -= parsedDuration.Milliseconds()
				} else {
					break
				}
			}

			//  некий алгоритм нахождения среднего значения
			avgValue = sumOfAllDatas / float64(listOfDataForLastThirtySeconds.Len())
			println(dataAndTime.data, " - recieved value")
		case <-ticker.C:
			fmt.Println("Current awg: ", avgValue)
		case <-interruptSignal:
			fmt.Printf("your AVG value is: %v , i`m ending \n", avgValue)
			os.Exit(1)
		}
	}
}
