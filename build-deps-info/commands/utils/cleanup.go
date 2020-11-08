package utils

func Cleanup(closeReader func() error, err *error) {
	er := closeReader()
	if er != nil && err == nil {
		*err = er
	}
}
