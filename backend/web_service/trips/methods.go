package trips

func (se StatusError) Error() string {
	return se.Error()
}

func (se StatusError) Unwrap() error {
	return se.Err
}
