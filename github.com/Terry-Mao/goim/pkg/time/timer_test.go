package time

import (
	"fmt"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	timer := NewTimer(2)


	timer.Add(time.Duration(time.Second*8), func() {
		fmt.Println("回调函数1")
	})
	timer.Add(time.Duration(time.Second*9), func() {
		fmt.Println("回调函数2")
	})
	timer.Add(time.Duration(time.Second*10), func() {
		fmt.Println("回调函数3")
	})

	timer.Add(time.Duration(time.Second*11), func() {
		fmt.Println("回调函数4")
	})
	timer.Add(time.Duration(time.Second*12), func() {
		fmt.Println("回调函数5")
	})
	timer.Add(time.Duration(time.Second*13), func() {
		fmt.Println("回调函数6")
	})



	time.Sleep(time.Second*40)


	//for i := 0; i < 2; i++ {
	//	//fmt.Println("time.Duration(i)",time.Duration(i))
	//	//fmt.Println("time.Second",time.Second)
	//	//fmt.Println("5*time.Minute",5*time.Minute)
	//	fmt.Println("xx:",time.Duration(i)*time.Second+5*time.Minute)
	//	tds[i] = timer.Add(time.Duration(i)*time.Second+5*time.Minute, func() {
	//		fmt.Println("回调函数")
	//	})
	//}




	//printTimer(timer)
	//for i := 0; i < 2; i++ {
	//	//fmt.Printf("td: %s, %s, %d", tds[i].Key, tds[i].ExpireString(), tds[i].index)
	//	timer.Del(tds[i])
	//}
	//printTimer(timer)
	//for i := 0; i < 2; i++ {
	//	tds[i] = timer.Add(time.Duration(i)*time.Second+5*time.Minute, nil)
	//}
	//printTimer(timer)
	//for i := 0; i < 2; i++ {
	//	timer.Del(tds[i])
	//}
	//printTimer(timer)
	//println("---------------------------")
	//timer.Add(time.Second, nil)
	//time.Sleep(time.Second * 2)
	//
	//fmt.Println(len(timer.timers))
	//if len(timer.timers) != 0 {
	//	t.FailNow()
	//}
}

func printTimer(timer *Timer) {
	//
	//type TimerData struct {
	//	Key    string
	//	expire itime.Time
	//	fn     func()
	//	index  int
	//	next   *TimerData
	//}
	//
	fmt.Println("TimerData 长度:", len(timer.timers))
	for i := 0; i < len(timer.timers); i++ {
		fmt.Println("key",timer.timers[i].Key)
		fmt.Println("ExpireString",timer.timers[i].ExpireString())
		fmt.Println("index",timer.timers[i].index)

		//fmt.Printf("timer: %s, %s, index: %d", timer.timers[i].Key, timer.timers[i].ExpireString(), timer.timers[i].index)
	}
}
