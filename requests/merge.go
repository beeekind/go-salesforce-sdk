package requests

import (
	"bytes"
)

// MergeJSONArrays takes many potentially large JSON arrays and merges them into a single array
// that can be unmarshalled into a single slice. Because of the performance impact of
// iterating through the entirety of an array we attempt to operate only on the first and last
// characters of the array.
//
// This method is used to create better API signatures when bulk querying - where methods further down
// the call stack are re-used and don't know the specific type of the response but need to combine them anyways.
//
// The same result can probably be achieved with less effort via the reflect package but this solution
// is succinct and works well enough.
//
// There is also room to optimize memory usage by pre-computing the length of the final output.
func MergeJSONArrays(arrays ...[]byte) []byte {
	buffer := bytes.NewBufferString("[")
	for i := 0; i < len(arrays); i++ {
		//
		if len(arrays[i]) == 0 {
			continue
		}

		arrays[i] = bytes.TrimPrefix(arrays[i], []byte("["))
		arrays[i] = bytes.TrimSuffix(arrays[i], []byte("]"))
		buffer.Write(arrays[i])

		// ensure valid JSON by comma delimiting the last joined object of this array
		//
		// not needed if its the last object in the final array
		if i != len(arrays)-1 {
			buffer.WriteString(",")
		}
	}

	buffer.WriteString("]")
	return buffer.Bytes()
}
