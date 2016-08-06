package webapps

import (
	"fmt"
	"sync"

	termbox "github.com/nsf/termbox-go"
)

// Screen represents a display to the current terminal
type Screen struct {
	Bg           termbox.Attribute
	FgBorder     termbox.Attribute
	FgLabels     termbox.Attribute
	marginTop    int
	marginRight  int
	marginBottom int
	marginLeft   int

	Done     chan struct{}
	Resized  chan struct{}
	Save     chan struct{}
	Clear    chan struct{}
	Generate chan struct{}

	mu sync.Mutex
}

// NewScreen instanciates a Screen object
func NewScreen() (*Screen, error) {
	s := &Screen{
		Bg:           termbox.ColorBlack,
		FgBorder:     termbox.ColorWhite,
		FgLabels:     termbox.ColorCyan,
		marginTop:    1,
		marginRight:  1,
		marginBottom: 2,
		marginLeft:   1,

		Done:     make(chan struct{}),
		Resized:  make(chan struct{}, 10),
		Save:     make(chan struct{}),
		Clear:    make(chan struct{}),
		Generate: make(chan struct{}),
	}
	err := s.init()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Screen) init() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	s.redrawDisplay()

	e := captureEvents(s)
	processEvents(s, e)

	return nil
}

// Close releases resources and terminal
func (s *Screen) Close() {
	termbox.Interrupt()
	<-s.Done
	termbox.Close()
}

func (s *Screen) redrawDisplay() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	termbox.Clear(s.FgBorder, s.Bg)

	err := s.drawBorder()
	if err != nil {
		return err
	}
	return s.drawControls()
}

func (s *Screen) drawBorder() error {
	w, h := termbox.Size()
	for x := s.marginLeft; x < w-s.marginRight; x++ {
		termbox.SetCell(x, -1+s.marginTop, '\u2500', s.FgBorder, s.FgBorder)
		termbox.SetCell(x, h-s.marginBottom, '\u2500', s.FgBorder, s.FgBorder)
	}
	for y := s.marginTop; y < h-s.marginBottom; y++ {
		termbox.SetCell(-1+s.marginLeft, y, '\u2502', s.FgBorder, s.FgBorder)
		termbox.SetCell(w-s.marginRight, y, '\u2502', s.FgBorder, s.FgBorder)
	}
	termbox.SetCell(s.marginLeft-1, s.marginTop-1, '\u250C', s.FgBorder, s.FgBorder)
	termbox.SetCell(s.marginLeft-1, h-s.marginBottom, '\u2514', s.FgBorder, s.FgBorder)
	termbox.SetCell(w-s.marginRight, h-s.marginBottom, '\u2518', s.FgBorder, s.FgBorder)
	termbox.SetCell(w-s.marginRight, s.marginTop-1, '\u2510', s.FgBorder, s.FgBorder)
	return termbox.Flush()
}

func (s *Screen) drawControls() error {
	_, h := termbox.Size()

	msg := "q :Quit  c :Clear IncludeList  s :Save IncludeList  g :Generate static assets"
	for i, c := range msg {
		termbox.SetCell(1+i, h-1, c, s.FgLabels, s.Bg)
	}
	return termbox.Flush()
}

func (s *Screen) displayMsg(msg string) error {
	s.clearMsg()
	w, h := termbox.Size()
	max := len(msg)
	if max > w-s.marginRight-s.marginLeft-2 {
		max = w - s.marginRight - s.marginLeft - 2
	}
	for x := 0; x < max; x++ {
		termbox.SetCell(x+s.marginLeft, h-s.marginBottom-1, rune(msg[x]), s.FgLabels, s.Bg)
	}
	return termbox.Flush()
}

func (s *Screen) clearMsg() error {
	w, h := termbox.Size()

	for x := s.marginLeft; x < w-s.marginRight-1; x++ {
		termbox.SetCell(x, h-s.marginBottom-1, ' ', s.FgLabels, s.Bg)
	}
	return termbox.Flush()
}

func (s *Screen) updateStats(list []string) error {
	_, h := termbox.Size()

	msg := fmt.Sprintf("%d files in the list.", len(list))
	for i, c := range msg {
		termbox.SetCell(1+i, h-s.marginBottom-2, c, s.FgLabels, s.Bg)
	}
	return termbox.Flush()
}

func captureEvents(s *Screen) <-chan termbox.Event {
	events := make(chan termbox.Event)
	go func() {
		defer close(events)
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventInterrupt:
				return
			default:
				select {
				case events <- ev:
				case <-s.Done:
				}
			}
		}
	}()
	return events
}

func processEvents(screen *Screen, events <-chan termbox.Event) {
	go func() {
		defer close(screen.Done)
		for {
			if ev, ok := <-events; ok {
				switch ev.Type {
				case termbox.EventKey:
					if ev.Ch == 0 {
						switch ev.Key {
						case termbox.KeyEsc, termbox.KeyCtrlQ:
							return
						}
					} else {
						switch ev.Ch {
						case 115, 83: // s, S
							screen.Save <- struct{}{}
						case 99, 67: // c, C
							screen.Clear <- struct{}{}
						case 103, 71: // g, G
							screen.Generate <- struct{}{}
						case 81, 113: // Q, q
							return
						default:
							// fmt.Println(ev.Ch)
						}
					}
				case termbox.EventResize:
					screen.redrawDisplay()
					screen.Resized <- struct{}{}
				}
			} else {
				return
			}
		}
	}()
}
