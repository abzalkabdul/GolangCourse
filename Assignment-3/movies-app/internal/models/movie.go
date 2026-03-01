package models

type Technician struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Movie struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Genre       string       `json:"genre"`
	Budget      int64        `json:"budget"`
	Technicians []Technician `json:"technicians,omitempty"`
}
