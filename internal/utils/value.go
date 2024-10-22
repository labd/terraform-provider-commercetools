package utils

// GetRef gets the reference of value with generic type
func GetRef[T any](value T) *T {
	return &value
}
