package main

import (
	"example/timewheel"
	"fmt"
	"time"
)

func main() {
	var now time.Time
	tw, err := timewheel.New(3*time.Second, 10, func(i interface{}) {
		fmt.Println(i, time.Now().Sub(now))
	})
	if err != nil {
		fmt.Println(err)
	}

	tw.Start()
	now = time.Now()

	err = tw.AddTimer(11*time.Second, nil, fmt.Sprintf("21"))
	if err != nil {
		fmt.Println(err)
	}
	err = tw.AddTimer(3*time.Second, nil, fmt.Sprintf("3"))
	if err != nil {
		fmt.Println(err)
	}
	err = tw.AddTimer(2*time.Second, nil, fmt.Sprintf("2"))
	if err != nil {
		fmt.Println(err)
	}
	err = tw.AddTimer(10*time.Second, nil, fmt.Sprintf("20"))
	if err != nil {
		fmt.Println(err)
	}
	err = tw.AddTimer(9*time.Second, nil, fmt.Sprintf("19"))
	if err != nil {
		fmt.Println(err)
	}
	select {}
}
