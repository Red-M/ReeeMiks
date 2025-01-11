package deej

// SliderMoveEvent represents a single slider move captured by deej
type SliderMoveEvent struct {
	SliderID     int
	PercentValue float32
}

type ButtonEvent struct {
	ButtonID     int
	Value 			 int
}

type DeejConnection interface {
	Start() error
	Stop()
	SubscribeToSliderMoveEvents() chan SliderMoveEvent
	SubscribeToButtonEvents() chan ButtonEvent
}
