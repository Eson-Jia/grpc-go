package resolver

type Node struct {
	ID     string `json:"id,omitempty"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

type Service struct {
	Name  string  `json:"name,omitempty"`
	Nodes []*Node `json:"nodes,omitempy"`
}
