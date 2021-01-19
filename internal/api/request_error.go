package api

type requestError struct {
	wrappedError error
	message      string
	statusCode   int
}

func (e requestError) Error() string {
	return e.message
}

func (e requestError) Unwrap() error {
	return e.wrappedError
}
