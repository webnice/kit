package ico

// FormatError Ошибка формата, вводимые данные не являются действительными для ICO.
type FormatError string

// Error Возврат ошибки.
func (e FormatError) Error() string { return "invalid ICO format: " + string(e) }
