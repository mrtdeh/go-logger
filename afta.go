package logger

import (
	"net/http"
	"reflect"
)

func (l *Logger) AFTA() *AftaLogger {
	return &AftaLogger{logger: &Logger{Logger: l.With().Str("type", "afta").Logger(), aftaLevel: l.aftaLevel}}
}

type AftaLogger struct {
	logger *Logger
}

type AftaStatement struct {
	logger *Logger
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
	return &AftaStatement{
		logger: &Logger{
			Logger: l.logger.With().
				Str("entity", entity).
				Any("record_id", id).
				Logger(),
		},
		action: "CREATE",
	}
}

func (l *AftaLogger) Update(entity string, id any) *AftaStatement {
	return &AftaStatement{
		logger: &Logger{
			Logger: l.logger.With().
				Str("entity", entity).
				Any("record_id", id).
				Logger(),
		},
		action: "UPDATE",
	}
}

func (l *AftaLogger) Message(msg string) *AftaStatement {
	return &AftaStatement{
		logger: &Logger{
			Logger: l.logger.With().
				Logger(),
		},
		msg:    msg,
		action: "MESSAGE",
	}
}

func (l *AftaLogger) Delete(entity string, id any) *AftaStatement {
	return &AftaStatement{
		logger: &Logger{
			Logger: l.logger.With().
				Str("entity", entity).
				Any("record_id", id).
				Logger(),
		},
		action: "DELETE",
	}
}

func (stm *AftaStatement) WithChanges(fieldsChange ...FieldChange) *AftaStatement {
	var arr []FieldChange
	arr = append(arr, fieldsChange...)
	stm.logger.Logger = stm.logger.With().Interface("changes", arr).Logger()
	return stm
}

func (stm *AftaStatement) WithModel(model interface{}) *AftaStatement {
	var arr []FieldChange

	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)

		arr = append(arr, FieldChange{
			Field:    field.Name,
			NewValue: fieldValue.Interface(),
		})
	}

	stm.logger.Logger = stm.logger.With().Interface("changes", arr).Logger()
	return stm
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
