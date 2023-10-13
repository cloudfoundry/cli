package actionerror

type TaskFailedError struct{}

func (TaskFailedError) Error() string {
	return "Task failed to complete successfully"
}
