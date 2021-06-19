package sampling

// Sample is a single sample data point. As Go currently has no generics the common
// pointer to all types interfac{} was used.
type Sample interface{}

// Sampler is an interface for sampling data from a stream.
type Sampler interface {

	// Add adds multiple samples to the Sampler. Returns all not accepted samples which
	// is a subset of `sample`
	Add(samples []Sample) []Sample

	// Data returns a slice of the current samples within the Sampler.
	Data() []Sample

	// Dropped returns the discarded samples after the last Add operation.
	Dropped() []Sample
}
