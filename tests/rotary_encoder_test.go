package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toxygene/periphio-ky-040-rotary-encoder/device"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpiotest"
)

func TestRotaryEncoder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClockPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
		mockDataPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
		mockSwitchPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}

		assert.NoError(t, mockClockPin.In(gpio.PullDown, gpio.RisingEdge))
		assert.NoError(t, mockDataPin.In(gpio.PullDown, gpio.NoEdge))
		assert.NoError(t, mockSwitchPin.In(gpio.PullDown, gpio.RisingEdge))

		rotaryEncoder := device.NewRotaryEncoder(
			mockClockPin,
			mockDataPin,
			mockSwitchPin,
			time.Millisecond,
		)

		ctx, cancel := context.WithCancel(context.Background())

		actions := make(chan device.Action)

		go func() {
			defer close(actions)
			defer cancel()

			mockDataPin.L = gpio.Low
			mockClockPin.L = gpio.High

			mockClockPin.EdgesChan <- gpio.High

			assert.Equal(t, device.Clockwise, <-actions)

			mockDataPin.L = gpio.Low
			mockClockPin.L = gpio.High

			mockClockPin.EdgesChan <- gpio.High

			assert.Equal(t, device.Clockwise, <-actions)

			mockDataPin.L = gpio.High
			mockClockPin.L = gpio.High

			mockClockPin.EdgesChan <- gpio.High

			assert.Equal(t, device.CounterClockwise, <-actions)

			mockSwitchPin.L = gpio.High

			mockSwitchPin.EdgesChan <- gpio.High

			assert.Equal(t, device.Click, <-actions)
		}()

		assert.NoError(t, rotaryEncoder.Run(ctx, actions))
	})
}
