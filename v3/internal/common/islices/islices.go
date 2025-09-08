// Internal package with generic helper functions for slices
package islices

// Transform applies a transformation function to each element of the input slice and returns a new slice with the transformed elements.
func Transform[In any, Out any](in []In, transform func(In) Out) []Out {
	out := make([]Out, 0, len(in))
	for _, item := range in {
		out = append(out, transform(item))
	}
	return out
}
