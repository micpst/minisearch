package store

import "fmt"

type DocumentNotFoundError struct {
	Id string
}

type DocumentAlreadyExistsError struct {
	Id string
}

type WrongSearchPropertyType struct {
	Property string
}

func (e *DocumentNotFoundError) Error() string {
	return fmt.Sprintf("Document with id '%s' not found", e.Id)
}

func (e *DocumentAlreadyExistsError) Error() string {
	return fmt.Sprintf("Document with id '%s' already exists", e.Id)
}

func (e *WrongSearchPropertyType) Error() string {
	return fmt.Sprintf("Property '%s' is not searchable", e.Property)
}
