package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toxygene/periphio-ky-040-rotary-encoder/device"
	"golang.org/x/sync/errgroup"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func main() {
	clockPinName := flag.String("clock", "", "clock pin name")
	dataPinName := flag.String("data", "", "data pin name")
	switchPinName := flag.String("switch", "", "switch pin name")
	timeout := flag.String("timeout", "1s", "timeout")

	flag.Parse()

	if *clockPinName == "" || *dataPinName == "" || *switchPinName == "" {
		flag.PrintDefaults()
		return
	}

	if _, err := host.Init(); err != nil {
		panic(err)
	}

	clockPin := gpioreg.ByName(*clockPinName)
	if err := clockPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		panic(err)
	}

	dataPin := gpioreg.ByName(*dataPinName)
	if err := dataPin.In(gpio.PullDown, gpio.NoEdge); err != nil {
		panic(err)
	}

	switchPin := gpioreg.ByName(*switchPinName)
	if err := switchPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		panic(err)
	}

	var rotaryEncoder *device.RotaryEncoder
	if *timeout == "" {
		rotaryEncoder = device.NewRotaryEncoder(
			clockPin,
			dataPin,
			switchPin,
			-1,
		)
	} else {
		timeoutDuration, err := time.ParseDuration(*timeout)
		if err != nil {
			panic(err)
		}

		rotaryEncoder = device.NewRotaryEncoder(
			clockPin,
			dataPin,
			switchPin,
			timeoutDuration,
		)
	}

	g, ctx := errgroup.WithContext(context.Background())

	actions := make(chan device.Action)

	g.Go(func() error {
		err := rotaryEncoder.Run(ctx, actions)

		close(actions)

		return err
	})

	g.Go(func() error {
		counter := 0

		for action := range actions {
			if action == device.Clockwise {
				counter++

				if counter == 14 {
					counter = 15
				} else if counter == 17 {
					counter = 0
				}

				println(counter)
			} else if action == device.CounterClockwise {
				counter--

				if counter == 14 {
					counter = 13
				} else if counter == -1 {
					counter = 16
				}

				println(counter)
			} else if action == device.Click {
				fmt.Printf("printing value %d\n", counter)
			}
		}

		return nil
	})

	g.Go(func() error {
		signals := make(chan os.Signal, 1)

		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

		<-signals

		return fmt.Errorf("process shutdown requested")
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
