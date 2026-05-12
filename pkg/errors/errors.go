package errors

func (e *DiffSummaryError) Error() string {
	return e.Message
}
