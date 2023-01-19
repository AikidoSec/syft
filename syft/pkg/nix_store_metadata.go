package pkg

type NixStoreMetadata struct {
	// Hash is the prefix of the nix store basename path
	Hash string `mapstructure:"hash" json:"hash"`

	// Output allows for optionally specifying the specific nix package output this package represents (for packages that support multiple outputs)
	Output string `mapstructure:"output" json:"output,omitempty"`
}