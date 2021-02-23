package chaincfg

// Network defines a type alias to represent the network parameters for
// various currencies.
type Network = string

// BitcoinNetwork defines the network (and therefore the chain parameters)
// that a Bitcoin keychain is associated to.
type BitcoinNetwork = Network

// LitecoinNetwork defines the network (and therefore the chain parameters)
// that a Litecoin keychain is associated to.
type LitecoinNetwork = Network
