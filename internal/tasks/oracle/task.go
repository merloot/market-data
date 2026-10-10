package oracle

const (
	Type = "market-data-oracle"
	Spec = "*/2 * * * *" // every 2m
)

type Payload struct {
}
