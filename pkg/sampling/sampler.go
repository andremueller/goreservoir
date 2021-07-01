package sampling

// Sample is a single sample data point. As Go currently has no generics the common
// pointer to all types interface{} was used.
type Sample interface{}

// Sampler is an interface for sampling data from a stream.
type Sampler interface {
	// Add adds one or multiple samples to the Sampler.
	// Returns a set of dropped samples in this step
	Add(samples []Sample) []Sample

	// Data returns a slice of the current samples within the Sampler.
	Data() []Sample
}
