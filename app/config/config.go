package config

type Server struct {
	ControlAddr string
	ClientAddrs []string
	UserAddrs   []string
}

type Client struct {
	ControlAddr string
	Proxys      []struct {
		Remote string
		Local  string
	}
}
