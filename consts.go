package econet

const HuwTemp = 1281

type BoilerStatus uint32

const (
	TurnedOff BoilerStatus = iota
	FireUp1
	FireUp2
	Work
	Supervision
	Halted
	Stop
	BurningOff
	Manual
	Alarm
	Unsealing
	Chimney
	Stabilization
	NoTransmission
)
