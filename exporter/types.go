package exporter

type Pricing struct {
	Product     Product
	ServiceCode string
	Terms       Terms
}

type Terms struct {
	OnDemand map[string]SKU
	Reserved map[string]SKU
}
type Product struct {
	ProductFamily string
	Attributes    map[string]string
	Sku           string
}

type SKU struct {
	PriceDimensions map[string]Details
	Sku             string
	EffectiveDate   string
	OfferTermCode   string
	TermAttributes  string
}

type Details struct {
	Unit         string
	EndRange     string
	Description  string
	AppliesTo    []string
	RateCode     string
	BeginRange   string
	PricePerUnit map[string]string
}
