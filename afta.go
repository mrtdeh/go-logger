package logger

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type AftaLogger struct {
	logger *MyLogger
}

type AftaStatement struct {
	logger *MyLogger
	action string
	msg    string
}

type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value,omitempty"`
	NewValue any    `json:"new_value"`
}

func (l *AftaLogger) NewFieldChange(field string, oldVal, newVal any) FieldChange {
	return FieldChange{Field: field, OldValue: oldVal, NewValue: newVal}
}

func (l *AftaLogger) Create(entity string, id any) *AftaStatement {
	s := &AftaStatement{
		logger: l.logger,
		action: "CREATE",
	}
	s.logger.Logger = l.logger.With().
		Str("entity", entity).
		Any("record_id", id).
		Logger()
	return s
}

func (l *AftaLogger) Update(entity string, id any) *AftaStatement {
	s := &AftaStatement{
		logger: l.logger,
		action: "UPDATE",
	}
	s.logger.Logger = l.logger.With().
		Str("entity", entity).
		Any("record_id", id).
		Logger()

	return s
}

func (l *AftaLogger) Message(msg string) *AftaStatement {
	s := &AftaStatement{
		logger: l.logger,
		msg:    msg,
		action: "MESSAGE",
	}
	s.logger.Logger = l.logger.With().Logger()
	return s
}

func (l *AftaLogger) Delete(entity string, id any) *AftaStatement {
	s := &AftaStatement{
		logger: l.logger,
		action: "DELETE",
	}
	s.logger.Logger = l.logger.With().
		Str("entity", entity).
		Any("record_id", id).
		Logger()

	return s
}

func (stm *AftaStatement) WithChanges(fieldsChange ...FieldChange) *AftaStatement {
	var arr []FieldChange
	arr = append(arr, fieldsChange...)
	stm.logger.Logger = stm.logger.With().Interface("changes", arr).Logger()
	return stm
}

func (stm *AftaStatement) WithModel(models ...interface{}) *AftaStatement {
	if len(models) == 0 || len(models) > 2 {
		panic(fmt.Sprintf("models count is invalid: %d", len(models)))
	}

	var changes []FieldChange

	newModel := models[0]
	newVal := reflect.ValueOf(newModel)
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	if len(models) == 2 && models[1] != nil {
		oldModel := models[1]
		oldVal := reflect.ValueOf(oldModel)
		if oldVal.Kind() == reflect.Ptr {
			oldVal = oldVal.Elem()
		}

		for i := 0; i < newVal.NumField(); i++ {
			// fmt.Println("i : ", i)
			field := newVal.Type().Field(i)
			newFieldValue := newVal.Field(i).Interface()
			oldFieldValue := oldVal.Field(i).Interface()

			// if newFieldValue != oldFieldValue {
			if !CompareTwoValues(newFieldValue, oldFieldValue) {
				changes = append(changes, FieldChange{
					Field:    field.Name,
					OldValue: filterFields(oldFieldValue),
					NewValue: filterFields(newFieldValue),
				})
			}
		}
	} else {
		for i := 0; i < newVal.NumField(); i++ {
			field := newVal.Type().Field(i)
			newFieldValue := newVal.Field(i).Interface()

			changes = append(changes, FieldChange{
				Field:    field.Name,
				NewValue: filterFields(newFieldValue),
			})
		}
	}

	stm.logger.Logger = stm.logger.With().Interface("changes", changes).Logger()
	return stm
}

func filterFields(st interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.ValueOf(st)
	typ := reflect.TypeOf(st)

	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)

			if tag, ok := field.Tag.Lookup("afta"); ok {
				if tag == "-" {
					continue
				} else if strings.HasPrefix(tag, "ref:") {
					refFieldName := strings.TrimPrefix(tag, "ref:")
					refField := val.FieldByName(refFieldName)
					if refField.IsValid() {
						result[field.Name] = refField.Interface()
					}
					continue
				}
			}

			result[field.Name] = fieldValue.Interface()
		}
	}
	return result
}

func CompareTwoValues(a, b interface{}) bool {
	// بررسی نوع داده‌ها
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	// استفاده از reflect.DeepEqual برای مقایسه عمیق
	return reflect.DeepEqual(a, b)
}

func (stm *AftaStatement) WithRequestHeader(header http.Header) *AftaStatement {
	var ip, user string
	if v, ok := header["User"]; ok {
		user = v[0]
	}
	if v, ok := header["Ip"]; ok {
		ip = v[0]
	}
	stm.logger.Logger = stm.logger.With().Str("user", user).Str("ip", ip).Logger()
	return stm
}

func (stm *AftaStatement) WithRecordID(id any) *AftaStatement {
	stm.logger.Logger = stm.logger.With().Any("record_id", id).Logger()
	return stm
}

func (stm *AftaStatement) Log() {
	if stm.logger.mock {
		return
	}

	if !stm.logger.aftaLevel {
		return
	}

	var msg string
	switch stm.action {
	case "CREATE":
		msg = "Record created"
	case "UPDATE":
		msg = "Record updated"
	case "MESSAGE":
		msg = stm.msg
	case "DELETE":
		msg = "Record deleted"
	default:
		msg = "Action performed"
	}

	stm.logger.Info().Str("action", stm.action).Msg(msg)
}
