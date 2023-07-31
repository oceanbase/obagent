package agentd

type Limiter interface {
	LimitPid(pid int) error
}

func NewLimiter(name string, conf LimitConfig) (Limiter, error) {
	return newLimiter(name, conf)
}

type NopLimiter struct {
}

func (l *NopLimiter) LimitPid(pid int) error {
	return nil
}
