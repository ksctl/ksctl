package e2e

type Operation string

const (
	OpCreate    = Operation("create")
	OpScaleUp   = Operation("scaleup")
	OpDelete    = Operation("delete")
	OpScaleDown = Operation("scaledown")
	OpGet       = Operation("get")
	OpSwitch    = Operation("switch")
)
