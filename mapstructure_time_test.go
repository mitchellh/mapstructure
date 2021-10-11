package mapstructure

import (
	"fmt"
	"time"
)


const (
	timeFormart = "2006-01-02 15:04:05"
)

func ExampleTimeDecode() {
	type TimeModel struct {
		AppointTime time.Time `mapstructure:"appointTime"`
	}

	t,_ := time.Parse(timeFormart,"2021-10-11 13:21:00")

	input := map[string]interface{}{
		"appointTime": t,
	}

	var model TimeModel

	err := Decode(input, &model)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", model)
	fmt.Println()
	fmt.Printf("%#v", input)

	//OutPut:
	//mapstructure.TimeModel{AppointTime:time.Date(2021, time.October, 11, 13, 21, 0, 0, time.UTC)}
	//map[string]interface {}{"appointTime":time.Date(2021, time.October, 11, 13, 21, 0, 0, time.UTC)}

}



func ExampleCustomDecode() {
	type TimeType time.Time
	type TimeModel struct {
		AppointTime TimeType `mapstructure:"appointTime"`
	}

	t,_ := time.Parse(timeFormart,"2021-10-11 13:21:00")

	input := map[string]interface{}{
		"appointTime": TimeType(t),
	}

	var model TimeModel

	err := Decode(input, &model)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", model.AppointTime)
	fmt.Println()
	fmt.Printf("%#v", input["appointTime"])

	//OutPut:
	//mapstructure.TimeType{wall:0x0, ext:63769555260, loc:(*time.Location)(nil)}
	//mapstructure.TimeType{wall:0x0, ext:63769555260, loc:(*time.Location)(nil)}

}
