package backend

type discoveryApp struct {
	Name    string            `xml:"name,attr"`
	Actions []discoveryAction `xml:"action"`
}
