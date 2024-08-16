package poller

type Poller interface {
	Get(string, string) ([]string, error)
}
