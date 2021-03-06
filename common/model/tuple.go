package model

import (
	"context"
	"fmt"
	"time"

	"reflect"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
)

var reteCTXKEY = RetecontextKeyType{}

//Tuple is a runtime representation of a data tuple
type Tuple interface {
	GetTupleType() TupleType
	GetTupleDescriptor() *TupleDescriptor

	GetProperties() []string
	GetString(name string) (val string, err error)
	GetInt(name string) (val int, err error)
	GetLong(name string) (val int64, err error)
	GetDouble(name string) (val float64, err error)
	GetBool(name string) (val bool, err error)
	//GetDateTime(name string) time.Time
	GetKey() TupleKey
}

//MutableTuple mutable part of the tuple
type MutableTuple interface {
	Tuple
	SetString(ctx context.Context, name string, value string) (err error)
	SetInt(ctx context.Context, name string, value int) (err error)
	SetLong(ctx context.Context, name string, value int64) (err error)
	SetDouble(ctx context.Context, name string, value float64) (err error)
	SetBool(ctx context.Context, name string, value bool) (err error)
	//SetDatetime(ctx context.Context, name string, value time.Time) (err error)

	//will try to coerce value to the named property's type
	SetValue(ctx context.Context, name string, value interface{}) (err error)
	//SetValues(ctx context.Context, values map[string]interface{}) (err error)
}

type tupleImpl struct {
	tupleType TupleType
	tuples    map[string]interface{}
	key       TupleKey
	td        *TupleDescriptor
}

func NewTuple(tupleType TupleType, values map[string]interface{}) (mtuple MutableTuple, err error) {
	td := GetTupleDescriptor(tupleType)
	if td == nil {
		return nil, fmt.Errorf("Tuple descriptor not found [%s]", string(tupleType))
	}
	t := tupleImpl{}
	err = t.initTuple(td, values)
	if err != nil {
		return nil, err
	}
	return &t, err
}

func NewTupleWithKeyValues(tupleType TupleType, values ...interface{}) (mtuple MutableTuple, err error) {

	td := GetTupleDescriptor(tupleType)
	if td == nil {
		return nil, fmt.Errorf("Tuple descriptor not found [%s]", string(tupleType))
	}
	t := tupleImpl{}
	err = t.initTupleWithKeyValues(td, values...)
	if err != nil {
		return nil, err
	}
	return &t, err
}

func (t *tupleImpl) GetTupleType() TupleType {
	return t.tupleType
}

func (t *tupleImpl) GetTupleDescriptor() *TupleDescriptor {
	return t.td
}

func (t *tupleImpl) GetProperties() []string {
	keys := []string{}
	for k := range t.tuples {
		keys = append(keys, k)
	}
	return keys
}

func (t *tupleImpl) GetString(name string) (val string, err error) {
	err = t.chkProp(name)
	if err != nil {
		return "", err
	}
	//try to coerce the tuple value to a string
	v, err := data.CoerceToString(t.tuples[name])

	return v, err
}

func (t *tupleImpl) GetInt(name string) (val int, err error) {
	err = t.chkProp(name)
	if err != nil {
		return 0, err
	}
	//try to coerce the tuple value to an integer
	v, err := data.CoerceToInteger(t.tuples[name])

	return v, err
}

func (t *tupleImpl) GetLong(name string) (val int64, err error) {
	err = t.chkProp(name)
	if err != nil {
		return 0, err
	}
	//try to coerce the tuple value to a long
	v, err := data.CoerceToLong(t.tuples[name])

	return v, err
}

func (t *tupleImpl) GetDouble(name string) (val float64, err error) {
	err = t.chkProp(name)
	if err != nil {
		return 0, err
	}
	//try to coerce the tuple value to a double
	v, err := data.CoerceToDouble(t.tuples[name])

	return v, err
}

func (t *tupleImpl) GetBool(name string) (val bool, err error) {
	err = t.chkProp(name)
	if err != nil {
		return false, err
	}
	//try to coerce tuple value to a boolean
	v, err := data.CoerceToBoolean(t.tuples[name])

	return v, err
}
func (t *tupleImpl) SetString(ctx context.Context, name string, value string) (err error) {
	return t.validateAndCallListener(ctx, name, value)
}
func (t *tupleImpl) SetInt(ctx context.Context, name string, value int) (err error) {
	return t.validateAndCallListener(ctx, name, value)
}
func (t *tupleImpl) SetLong(ctx context.Context, name string, value int64) (err error) {
	return t.validateAndCallListener(ctx, name, value)
}
func (t *tupleImpl) SetDouble(ctx context.Context, name string, value float64) (err error) {
	return t.validateAndCallListener(ctx, name, value)
}
func (t *tupleImpl) SetBool(ctx context.Context, name string, value bool) (err error) {
	return t.validateAndCallListener(ctx, name, value)
}

func (t *tupleImpl) SetDatetime(ctx context.Context, name string, value time.Time) (err error) {
	return t.validateAndCallListener(ctx, name, value)
}

func (t *tupleImpl) SetValue(ctx context.Context, name string, value interface{}) (err error) {
	return t.validateAndCallListener(ctx, name, value)
}
func (t *tupleImpl) GetKey() TupleKey {
	return t.key
}

func (t *tupleImpl) initTuple(td *TupleDescriptor, values map[string]interface{}) (err error) {
	t.tuples = make(map[string]interface{})
	t.tupleType = TupleType(td.Name)
	t.td = td
	tk := tupleKeyImpl{}
	tk.keys = make(map[string]interface{})
	tk.td = *t.td
	t.key = &tk
	for _, tdp := range td.Props {
		val, found := values[tdp.Name]
		if found {
			coerced, err := data.CoerceToValue(val, tdp.PropType)
			if err == nil {
				t.tuples[tdp.Name] = coerced
			} else {
				return err
			}
		} else if tdp.KeyIndex != -1 { //key prop
			return fmt.Errorf("Key property [%s] not found", tdp.Name)
		}
	}

	for _, keyProp := range td.GetKeyProps() {
		tk.keys[keyProp] = t.tuples[keyProp]
	}
	return err
}

func (t *tupleImpl) initTupleWithKeyValues(td *TupleDescriptor, values ...interface{}) (err error) {
	t.tuples = make(map[string]interface{})
	t.tupleType = TupleType(td.Name)
	t.td = td
	tk := tupleKeyImpl{}
	tk.keys = make(map[string]interface{})
	tk.td = *t.td
	t.key = &tk
	if len(values) != len(t.td.GetKeyProps()) {
		return fmt.Errorf("Wrong number of key values in type [%s]. Expecting [%d], got [%d]",
			td.Name, len(t.td.GetKeyProps()), len(values))
	}

	i := 0
	for _, keyProp := range td.GetKeyProps() {
		tdp := td.GetProperty(keyProp)
		val := values[i]
		coerced, err := data.CoerceToValue(val, tdp.PropType)
		if err == nil {
			t.tuples[keyProp] = coerced
			tk.keys[keyProp] = coerced
		} else {
			return fmt.Errorf("Type mismatch for field [%s] in type [%s] Expecting [%s], got [%v]",
				keyProp, td.Name, tdp.PropType.String(), reflect.TypeOf(val))
		}
		i++
	}
	return err
}

func (t *tupleImpl) validateAndCallListener(ctx context.Context, name string, value interface{}) (err error) {

	if t.isKeyProp(name) {
		return fmt.Errorf("Cannot change a key property [%s] for type [%s]", name, t.td.Name)
	}

	err = t.validateNameValue(name, value)
	if err != nil {
		return err
	}
	if t.tuples[name] != value {
		t.tuples[name] = value
		callChangeListener(ctx, t, name)
	}
	return nil
}

func (t *tupleImpl) chkProp(name string) (err error) {
	//TODO: Check property's type and value's type compatibility
	prop := t.td.GetProperty(name)
	if prop != nil {
		return nil
	}
	return fmt.Errorf("Property [%s] undefined for type [%s]", name, t.td.Name)
}

func callChangeListener(ctx context.Context, tuple Tuple, prop string) {
	if ctx != nil {
		ctxR := ctx.Value(reteCTXKEY)
		if ctxR != nil {
			valChangeLister := ctxR.(ValueChangeListener)
			valChangeLister.OnValueChange(tuple, prop)
		}
	}
}

func (t *tupleImpl) validateNameValue(name string, value interface{}) (err error) {
	p := t.td.GetProperty(name)

	if p != nil {
		_, err := data.CoerceToValue(value, p.PropType)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("Property [%s] undefined for type [%s]", name, t.td.Name)
}

func (t *tupleImpl) isKeyProp(propName string) bool {
	found := false
	switch tki := t.key.(type) {
	case *tupleKeyImpl:
		_, found = tki.keys[propName]
	}
	return found
}
