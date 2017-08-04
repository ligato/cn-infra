package sql

import (
	"github.com/ligato/cn-infra/utils/safeclose"
	"reflect"
)

// SliceIt reads everything from the ValIterator & stores it to pointerToASlice
// It closes the iterator (since nothing left in the iterator)
func SliceIt(pointerToASlice interface{}, it ValIterator) error {
	/* TODO defer func() {
		if exp := recover(); exp != nil && it != nil {
			logger.Error(exp)
			exp = safeclose.Close(it)
			if exp != nil {
				logger.Error(exp)
			}
		}
	}()*/

	sl := reflect.ValueOf(pointerToASlice)
	if sl.Kind() == reflect.Ptr {
		sl = sl.Elem()
	} else {
		panic("must be pointer")
	}

	if sl.Kind() != reflect.Slice {
		panic("must be slice")
	}

	sliceType := sl.Type()

	sliceElemType := sliceType.Elem()
	sliceElemPtr := sliceElemType.Kind() == reflect.Ptr
	if sliceElemPtr {
		sliceElemType = sliceElemType.Elem()
	}
	for {
		row := reflect.New(sliceElemType)
		if stop := it.GetNext(row.Interface()); stop {
			break
		}

		if sliceElemPtr {
			sl.Set(reflect.Append(sl, row))
		} else {
			sl.Set(reflect.Append(sl, row.Elem()))
		}
	}

	return safeclose.Close(it)
}
