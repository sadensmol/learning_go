package main

import (
	"fmt"
	"time"
)

type budget struct {
	CampaingID string
	Balance float64
	Expires time.Time
}

func main() {
	b1:=budget{"compain_id", 12.11, time.Now()}

	fmt.Println(b1)
	b1.killTheDuck()
 }

 func (b budget) killTheDuck() {
	fmt.Println("kill the duck!")
 }