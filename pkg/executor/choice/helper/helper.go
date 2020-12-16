package helper

// Helper is an function help choice logic
type Helper func(variable string, targetString *string, targetInt *int, targetFloat *float64) (bool, error)
