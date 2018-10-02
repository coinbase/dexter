package engine

//
// A poller is anyy structure that can return a channel of investigations.
//
type Poller interface {
	Poll() chan Investigation
}
