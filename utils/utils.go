package utils

// Append a byte-decoded newline to a bytearray.
// This helper function exists only to add a descriptive name to this common
// operation.
func BinaryNewline(s []byte) []byte {
	return append(s, []byte{10}...)
}
