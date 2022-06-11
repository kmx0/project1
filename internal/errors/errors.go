package errors

import "errors"

//400 неверный формат запроса
var ErrStatusBadRequest = errors.New("неверный формат запроса")

// логин уже занят
var ErrStatusConflict = errors.New("логин уже занят")

//неверная пара логин/пароль
var ErrStatusUnauthorized = errors.New("неверная пара логин/пароль")
var ErrStatusInternalServer = errors.New("внутренняя ошибка сервера")
