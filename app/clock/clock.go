package clock

import "time"

type Clock interface{ Now() time.Time }

type System struct{}

func (System) Now() time.Time { return time.Now() }

type fixed struct{ t time.Time }

func (f fixed) Now() time.Time { return f.t }

// Fixed returns a Clock that always reports t (deterministic tests).
func Fixed(t time.Time) Clock { return fixed{t: t} }
