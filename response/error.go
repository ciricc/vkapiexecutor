package response

/* Объект ошибки, полученной в теле ответа HTTP запроса к API
   Обрабатывает только ошибки, относящиеся к выполнению API метода
*/
type Error struct {
	error
	message string
	intCode int
}

func NewError(message string, intCode int) *Error {
	return &Error{
		message: message,
		intCode: intCode,
	}
}

// Возвращает описание ошибки
func (v *Error) Error() string {
	return v.message
}

// Возвращает числовой код ошибки
func (v *Error) IntCode() int {
	return v.intCode
}
