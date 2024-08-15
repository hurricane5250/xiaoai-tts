package goxiaoai

import (
	"fmt"
	"testing"
)

func TestLogin(t *testing.T) {
	_, err := New("xxx", "xxx")
	if err != nil {
		panic(err)
	}

}

func TestGetDevice(t *testing.T) {
	x, err := New("xxx", "xxx")
	if err != nil {
		panic(err)
	}

	msg, err := x.GetDevices()
	if err != nil {
		panic(err)
	}

	fmt.Println(msg)
}

func TestLastAsk(t *testing.T) {
	x, err := New("xxx", "xxx")
	if err != nil {
		panic(err)
	}

	if err := x.GetLastAsk(); err != nil {
		panic(err)
	}

}
