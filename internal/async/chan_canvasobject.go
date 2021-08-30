// Code generated by go run gen.go; DO NOT EDIT.

package async

import "fyne.io/fyne/v2"

// UnboundedCanvasObjectChan is a channel with an unbounded buffer for caching
// CanvasObject objects.
//
// Delicate dance: One must aware that an unbounded channel may lead
// to OOM when the consuming speed of the buffer is lower than the
// producing speed constantly. However, such a channel may be fairly
// used for event delivering if the consumer of the channel consumes
// the incoming forever. Especially, rendering and even processing tasks.
type UnboundedCanvasObjectChan struct {
	in, out chan fyne.CanvasObject
}

// NewUnboundedCanvasObjectChan returns a unbounded channel with unlimited capacity.
func NewUnboundedCanvasObjectChan() *UnboundedCanvasObjectChan {
	ch := &UnboundedCanvasObjectChan{
		// The size of CanvasObject is less than 16-bit, we use 128 to fit
		// a CPU cache line (L2, 256 Bytes), which may reduce cache misses.
		in:  make(chan fyne.CanvasObject, 128),
		out: make(chan fyne.CanvasObject, 128),
	}
	go func() {
		// This is a preallocation of the internal unbounded buffer.
		// The size is randomly picked. But if one changes the size, the
		// reallocation size at the subsequent for loop should also be
		// changed too. Furthermore, there is no memory leak since the
		// queue is garbage collected.
		q := make([]fyne.CanvasObject, 0, 1<<10)
		for {
			e, ok := <-ch.in
			if !ok {
				close(ch.out)
				return
			}
			q = append(q, e)
			for len(q) > 0 {
				select {
				case ch.out <- q[0]:
					q = q[1:]
				case e, ok := <-ch.in:
					if ok {
						q = append(q, e)
						break
					}
					for _, e := range q {
						ch.out <- e
					}
					close(ch.out)
					return
				}
			}
			// If the remaining capacity is too small, we prefer to
			// reallocate the entire buffer.
			if cap(q) < 1<<5 {
				q = make([]fyne.CanvasObject, 0, 1<<10)
			}
		}
	}()
	return ch
}

// In returns a send-only channel that can be used to send values
// to the channel.
func (ch *UnboundedCanvasObjectChan) In() chan<- fyne.CanvasObject { return ch.in }

// Out returns a receive-only channel that can be used to receive
// values from the channel.
func (ch *UnboundedCanvasObjectChan) Out() <-chan fyne.CanvasObject { return ch.out }