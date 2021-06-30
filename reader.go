package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Reader struct {
	Header bool

	reader     *csv.Reader
	fieldindex []int
}

func NewReader(reader io.Reader) *Reader {
	r := csv.NewReader(reader)
	r.ReuseRecord = true
	return &Reader{reader: r}
}

func (r *Reader) Read(ref interface{}) error {
	rec, err := r.reader.Read()
	if err != nil {
		return err
	}

	val := reflect.ValueOf(ref)
	typ := val.Type()

	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("Cannot read into type %T: Expected pointer to struct", ref)
	}

	val = val.Elem()
	typ = val.Type()

	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("Cannot read into type %T: Expected pointer to struct", ref)
	}

	if r.fieldindex == nil {
		r.fieldindex = make([]int, len(rec))
		for ri := range r.fieldindex {
			r.fieldindex[ri] = -1
		}

		for i := 0; i < typ.NumField(); i++ {
			ftype := typ.Field(i)

			key := ftype.Name

			if tv, ok := ftype.Tag.Lookup("csv"); ok {
				t, _ := parseTag(tv)
				if t != "" {
					key = t
				}

				if t == "-" {
					continue
				}
			}

			if r.Header {
				for ri, rv := range rec {
					if strings.ToLower(rv) == strings.ToLower(key) {
						r.fieldindex[ri] = i
					}
				}
			} else if i < len(rec) {
				r.fieldindex[i] = i
			}
		}

		if r.Header {
			rec, err = r.reader.Read()
			if err != nil {
				return err
			}
		}
	}

	for ri, i := range r.fieldindex {
		if i != -1 {
			val.Field(i).Set(reflect.ValueOf(rec[ri]))
		}
	}

	return nil
}
