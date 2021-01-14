package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/ui"
)

// Height is the maximum height of the UI.
const Height = 18

// Those constants represent the generic colors we use in this example.
const (
	ColorYellow gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorMagenta
	ColorCyan
	ColorGreen
	ColorRed
)

type model struct {
	grid      gruid.Grid
	wpm       float64       // words per minute
	nlines    int           // number of lines per frame
	words     []string      // words from file
	par       *ui.Label     // text box
	help      *ui.Label     // help box
	info      *ui.Label     // info box
	pause     bool          // whether in pause or running
	i         int           // current word index
	frames    []int         // frames
	f         int           // current frame index
	gotoFrame int           // go to frame number
	wpf       float64       // words per frame (mean)
	wml       float64       // words length (mean)
	interval  time.Duration // time interval between two frames
}

func newModel(words []string) *model {
	gd := gruid.NewGrid(Width, Height)
	md := &model{
		grid:   gd,
		wpm:    250.0,
		nlines: 1,
		words:  words,
		frames: []int{}}
	st := gruid.Style{}
	md.par = &ui.Label{
		Box:        &ui.Box{Style: st.WithFg(ColorYellow)},
		StyledText: ui.StyledText{}.WithMarkup('r', st.WithFg(ColorRed)),
	}
	md.info = &ui.Label{
		Box:        &ui.Box{Title: ui.NewStyledText("Info: " + flag.Arg(0)), Style: st.WithFg(ColorMagenta)},
		StyledText: ui.StyledText{}.WithStyle(st.WithFg(ColorGreen)),
	}
	md.help = &ui.Label{
		Box:        &ui.Box{Title: ui.NewStyledText("Help"), Style: st.WithFg(ColorCyan)},
		StyledText: ui.NewStyledText(helpText).WithStyle(st.WithFg(ColorGreen)),
	}
	return md
}

const helpText = `+/-:speed  ;/,: inc/dec n° of lines  W/w: inc/dec n° of words
</>, (/), [/]:1, 50, 1000 backwards/forward
0-9*:frame number  g:goto frame  c:clear goto
p:pause  q:quit`

func (md *model) Update(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgInit:
		md.updateFrameInfo()
		return tick(md.interval)
	case timeMsg:
		if md.pause {
			break
		}
		if md.i > 0 {
			md.f++
		}
		var text string
		text, md.i = parText(md.words, md.i, md.nlines)
		md.par.SetText(text)
		if md.f >= len(md.frames)-1 {
			md.f = len(md.frames) - 1
			md.pause = true
		}
		md.updateInfo()
		wl := wlen(text)
		adjustment := float64(wl)/md.wpf - md.wml
		if adjustment > 5 {
			adjustment = 5
		} else if adjustment < -5 {
			adjustment = -5
		}
		return tick(md.interval + time.Duration(15*adjustment*float64(md.interval)/100))
	case gruid.MsgKeyDown:
		return md.updateMsgKeyDown(msg)
	}
	return nil
}

func (md *model) Draw() gruid.Grid {
	md.grid.Fill(gruid.Cell{Rune: ' '})
	rg := md.grid.Range()
	from := func(y int) gruid.Grid {
		return md.grid.Slice(rg.Lines(y, Height))
	}
	y := 0
	y += md.info.Draw(from(y)).Size().Y
	y += md.par.Draw(from(y)).Size().Y
	md.help.Draw(from(y))
	return md.grid
}

func (md *model) updateFrameInfo() {
	md.wpf, md.wml = md.frameInfo(md.nlines)
	md.interval = wpm2interval(md.wpm, md.words, md.wpf)
	for md.f = 0; md.f < len(md.frames) && md.frames[md.f] < md.i; md.f++ {
	}
	md.i = md.frames[md.f]
	md.updateInfo()
}

type timeMsg time.Time

func tick(d time.Duration) gruid.Cmd {
	t := time.NewTimer(d)
	return func() gruid.Msg {
		return timeMsg(<-t.C)
	}
}

func (md *model) frameInit() {
	text, _ := parText(md.words, md.i, md.nlines)
	md.par.SetText(text)
}

func (md *model) updateMsgKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	switch msg.Key {
	case "q", "Q", gruid.KeyEscape:
		return gruid.End()
	case "+":
		if md.wpm <= 950 {
			md.wpm += 50
			md.interval = wpm2interval(md.wpm, md.words, md.wpf)
			md.updateInfo()
		}
	case "-":
		if md.wpm >= 150 {
			md.wpm -= 50
			md.interval = wpm2interval(md.wpm, md.words, md.wpf)
			md.updateInfo()
		}
	case "W":
		if OptWords < 4 {
			OptWords++
			md.updateFrameInfo()
			md.frameInit()
		}
	case "w":
		if OptWords > 1 {
			OptWords--
			md.updateFrameInfo()
			md.frameInit()
		}
	case ";":
		if md.nlines < 3 {
			md.nlines++
			md.updateFrameInfo()
			md.frameInit()
		}
	case ",":
		if md.nlines > 1 {
			md.nlines--
			md.updateFrameInfo()
			md.frameInit()
		}
	case "g", gruid.KeyHome:
		if md.gotoFrame >= len(md.frames) {
			md.gotoFrame = len(md.frames) - 1
		}
		md.f = md.gotoFrame
		md.i = md.frames[md.f]
		md.gotoFrame = 0
		md.updateInfo()
		md.frameInit()
	case "c":
		md.gotoFrame = 0
	case "p", "P", gruid.KeySpace:
		if md.f < len(md.frames)-1 {
			md.pause = !md.pause
			if !md.pause {
				return tick(md.interval)
			}
		}
	case ">", gruid.KeyEnter, gruid.KeyArrowRight:
		md.frameNavFunc(1)
	case "<", gruid.KeyBackspace, gruid.KeyArrowLeft:
		md.frameNavFunc(-1)
	case ")", gruid.KeyPageDown:
		md.frameNavFunc(50)
	case "(", gruid.KeyPageUp:
		md.frameNavFunc(-50)
	case "]":
		md.frameNavFunc(1000)
	case "[":
		md.frameNavFunc(-1000)
	default:
		var n int
		_, err := fmt.Sscan(string(msg.Key), &n)
		if err == nil {
			md.gotoFrame = (md.gotoFrame * 10) + n
			if md.gotoFrame >= len(md.frames) {
				md.gotoFrame = len(md.frames) - 1
			}
			md.updateInfo()
		}
	}
	return nil
}

func (md *model) frameNavFunc(frames int) {
	switch {
	case md.f+frames >= len(md.frames):
		md.f = len(md.frames) - 1
	case md.f+frames < 0:
		md.f = 0
	default:
		md.f += frames
	}
	md.i = md.frames[md.f]
	md.updateInfo()
	md.frameInit()
}

func (md *model) updateInfo() {
	md.info.SetText(fmt.Sprintf(`frames/total: %d/%d
words/total: %d/%d  
words/frame: %.1f  wpm: ≈%.0f  interval: ≈%s  lines: %d
goto: %d`,
		md.f, len(md.frames)-1, md.i, len(md.words)-1, md.wpf, md.wpm, md.interval, md.nlines,
		md.gotoFrame))
}

func wpm2interval(wpm float64, words []string, wpf float64) time.Duration {
	du := int64(wpm / wpf)
	interval := time.Duration(wpm2ms(du)) * time.Millisecond

	return interval
}

func wpm2ms(wpm int64) int64 {
	return (60 * 1000) / wpm
}

// frameInfo updates md.frames and returns the mean number of words per frame,
// and the mean length of words.
func (md *model) frameInfo(n int) (float64, float64) {
	// n should be > 0
	i := 0
	steps := 0
	md.frames = md.frames[:0]
	wl := 0.0
	for i < len(md.words) {
		md.frames = append(md.frames, i)
		m := n
		for m > 0 {
			count, s, end := nextWords(md.words[i:])
			i += count
			wl += float64(wlen(s))
			m--
			if end {
				break
			}
		}
		steps++
	}
	return float64(len(md.words)) / float64(steps), wl / float64(len(md.words))
}

func alignCenter(s string, length int) string {
	s, index, l := highlight(s)
	if length < l {
		return s
	}
	b := strings.Builder{}
	d := length - l
	b.WriteString(strings.Repeat(" ", length/2-index))
	b.WriteString(s)
	b.WriteString(strings.Repeat(" ", d-length/2+index))
	return b.String()
}

func wlen(s string) int {
	c := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		c++
	}
	return c
}

func highlight(s string) (string, int, int) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, 0, 0
	}
	index := 0
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsSpace(r) {
			continue
		}
		index = i
		if index >= len(runes)/3 {
			break
		}
	}
	s = fmt.Sprintf("%s%s%c%s%s",
		string(runes[:index]), "@r", runes[index], "@N", string(runes[index+1:]))
	return s, index, len(runes)
}

func parText(words []string, i, n int) (string, int) {
	lines := make([]string, 0, n)
	for n > 0 && i < len(words) {
		count, next, end := nextWords(words[i:])
		i += count
		lines = append(lines, alignCenter(next, Width-2))
		n--
		if end {
			break
		}
	}
	return strings.Join(lines, "\n"), i
}

func nextWords(words []string) (int, string, bool) {
	b := strings.Builder{}
	count := 0
	end := false
	for i, w := range words {
		if i < OptWords && (b.Len() == 0 || b.Len()+len(w) < OptTextWidth) {
			if b.Len() > 0 {
				b.WriteRune(' ')
			}
			b.WriteString(w)
			last, _ := utf8.DecodeLastRuneInString(w)
			if unicode.IsPunct(last) {
				count = i + 1
				if endsByEndPunct(b.String()) {
					end = true
				}
				break
			}
		} else {
			count = i
			break
		}
	}
	return count, b.String(), end
}

func endsByEndPunct(s string) bool {
	end := false
	for len(s) > 0 {
		last, size := utf8.DecodeLastRuneInString(s)
		if !unicode.IsPunct(last) {
			return end
		}
		if last == '.' || last == '?' || last == '!' {
			end = true
		}
		s = s[:len(s)-size]
	}
	return end
}
