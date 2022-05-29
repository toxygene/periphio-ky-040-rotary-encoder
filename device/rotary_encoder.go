package device

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
	"periph.io/x/conn/v3/gpio"
)

type Action string

var (
	Clockwise        Action = "clockwise"
	CounterClockwise Action = "counter clockwise"
	Click            Action = "click"
)

type RotaryEncoder struct {
	clockPin  gpio.PinIO
	dataPin   gpio.PinIO
	switchPin gpio.PinIO
	timeout   time.Duration
}

func NewRotaryEncoder(clockPin gpio.PinIO, dataPin gpio.PinIO, switchPin gpio.PinIO, timeout time.Duration) *RotaryEncoder {
	return &RotaryEncoder{
		clockPin:  clockPin,
		dataPin:   dataPin,
		switchPin: switchPin,
		timeout:   timeout,
	}
}

func (r *RotaryEncoder) Run(ctx context.Context, actions chan<- Action) error {
	g, childCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		for {
			select {
			case <-childCtx.Done():
				return nil
			default:
				if !r.clockPin.WaitForEdge(r.timeout) {
					continue
				}

				if r.clockPin.Read() == gpio.High {
					if r.dataPin.Read() == gpio.Low {
						actions <- Clockwise
					} else {
						actions <- CounterClockwise
					}
				}
			}
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-childCtx.Done():
				return nil
			default:
				if !r.switchPin.WaitForEdge(time.Second) || r.switchPin.Read() == gpio.Low {
					continue
				}

				actions <- Click

				<-time.NewTimer(250 * time.Millisecond).C
			}
		}
	})

	g.Go(func() error {
		<-childCtx.Done()

		return r.clockPin.Halt()
	})

	g.Go(func() error {
		<-childCtx.Done()

		return r.switchPin.Halt()
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run rotary encoder: %w", err)
	}

	return nil
}
