package logtimer

import (
    "time"
)

type Timer struct {
    tick chan bool
    done chan bool
    seconds int
}

func New(seconds int) (Timer) {
    timer := Timer{
        tick: make(chan bool),
        done: make(chan bool),
        seconds: seconds,
    }
    go TimerLoop(timer.tick, timer.done, seconds)
    return timer
}

func (t Timer) Tick() (<-chan bool) {
    return t.tick
}

func (t Timer) Stop() {
    t.done <- true
}

func TimerLoop(tick chan<- bool, doneTimer <-chan bool, seconds int) {
    for {
        select {
        case <-time.After(time.Duration(seconds) * time.Second):
            tick <- true
        case <-doneTimer:
            return
        }
    }
}
