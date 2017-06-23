package safeclose

import (
	"errors"
	"io"
	"reflect"
	"github.com/prometheus/common/log"
)

// CloserWithoutErr is similar interface to GoLang Closer but Close() does not return error
type CloserWithoutErr interface {
	Close()
}

// Close closes closable I/O stream.
func  Close(obj interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered in safeclose", r)
		}
	}()

	if obj != nil {
		if closer, ok := obj.(*io.Closer); ok {
			if closer != nil {
				err := (*closer).Close()
				return err
			}
		} else if closer, ok := obj.(*CloserWithoutErr); ok {
			if closer != nil {
				(*closer).Close()
			}
		} else if closer, ok := obj.(io.Closer); ok {
			if closer != nil {
				log.Debug("closer: ", closer)
				err := closer.Close()
				return err
			}
		} else if closer, ok := obj.(CloserWithoutErr); ok {
			if closer != nil {
				closer.Close()
			}
		}
	}
	return nil
}

// CloseAll tries to close all objects and return all errors (there are nils if there was no errors).
func CloseAll(objs ...interface{}) (details []error, errOccured error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered in safeclose", r)
		}
	}()

	details = make([]error, len(objs))
	for i, obj := range objs {
		if obj != nil {
			if closer, ok := obj.(*io.Closer); ok {
				if closer != nil {
					err := (*closer).Close()
					if err != nil {
						details[i] = err
						errOccured = err

					}
				}
			} else if closer, ok := obj.(*CloserWithoutErr); ok {
				if closer != nil {
					(*closer).Close()
				}
			} else if closer, ok := obj.(io.Closer); ok {
				if closer != nil {
					err := closer.Close()
					if err != nil {
						details[i] = err
						errOccured = err

					}
				}
			} else if closer, ok := obj.(CloserWithoutErr); ok {
				if closer != nil {
					closer.Close()
				}
			} else if reflect.TypeOf(obj).Kind() == reflect.Chan {
				//reflect.ValueOf(nil).

				if x, ok := obj.(chan interface{}); ok {
					close(x)
				}
			}
		}
	}

	if errOccured != nil {
		return details, format(details)
	}

	return details, nil
}

// format squashes multiple errors into one.
func format(errs []error) error {
	errMsg := ""

	for _, err := range errs {
		if err != nil {
			errMsg += ";" + err.Error()
		}
	}

	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}
