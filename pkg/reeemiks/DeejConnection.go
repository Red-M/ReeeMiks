package reeemiks

// SliderMoveEvent represents a single slider move captured by reeemiks
type SliderMoveEvent struct {
	SliderID     int
	PercentValue float32
}

type ButtonEvent struct {
	ButtonID     int
	Value 			 int
}

type ReeemiksConnection interface {
	Start() error
	Stop()
	SubscribeToSliderMoveEvents() chan SliderMoveEvent
	SubscribeToButtonEvents() chan ButtonEvent
}
