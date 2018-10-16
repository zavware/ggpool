package ggpool

//TimeoutError is temporary error.
//This type of error can be catched by Temporary interface
type TimeoutError struct {
	err string
}

func (e *TimeoutError) Error() string {
	return e.err
}

//IsTemporary says that TimeoutError is Temporary error
func (e *TimeoutError) IsTemporary() bool {
	return true
}

//Temporary is interface for temporary pool errors.
//This type of error can be returned by Pool.Get() if Timeout (see Config) exceeded
type Temporary interface {
	IsTemporary() bool
}
