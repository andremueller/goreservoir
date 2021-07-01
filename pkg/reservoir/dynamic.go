package reservoir

// DynamicSampler is a biased dynamic sampler.
type DynamicSampler struct {
}

func NewDynamic() *DynamicSampler {
	return &DynamicSampler{}
}
