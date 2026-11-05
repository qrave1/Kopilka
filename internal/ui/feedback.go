package ui

import (
	"time"

	"fyne.io/fyne/v2"
)

// afterTap gives Fyne's press highlight time to appear before changing screens.
func (s *screen) afterTap(action func()) func() {
	return func() {
		if s.tapPending {
			return
		}
		if !fyne.CurrentApp().Settings().ShowAnimations() {
			action()
			return
		}

		s.tapPending = true
		current := s.window.Content()
		fyne.NewAnimation(90*time.Millisecond, func(progress float32) {
			if progress < 1 {
				return
			}
			fyne.Do(func() {
				s.tapPending = false
				if s.window.Content() == current {
					action()
				}
			})
		}).Start()
	}
}
