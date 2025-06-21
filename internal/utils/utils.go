package utils

func SendErrResponse(err error) map[string]string {
	return map[string]string{
		"error": err.Error(),
	}
}
