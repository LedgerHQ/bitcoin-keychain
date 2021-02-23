package integration

import pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"

type Fixture struct {
	ExternalDescriptor string
	InternalDescriptor string
	ExtendedPublicKey  string
	ChainParams        *pb.ChainParams
	Scheme             pb.Scheme
}

var BitcoinMainnetP2PKH = Fixture{
	ExternalDescriptor: "pkh(xpub6DCi5iJ57ZPd5qPzvTm5hUt6X23TJdh9H4NjNsNbt7t7UuTMJfawQWsdWRFhfLwkiMkB1rQ4ZJWLB9YBnzR7kbs9N8b2PsKZgKUHQm1X4or/0/*)",
	InternalDescriptor: "pkh(xpub6DCi5iJ57ZPd5qPzvTm5hUt6X23TJdh9H4NjNsNbt7t7UuTMJfawQWsdWRFhfLwkiMkB1rQ4ZJWLB9YBnzR7kbs9N8b2PsKZgKUHQm1X4or/1/*)",
	ExtendedPublicKey:  "xpub6DCi5iJ57ZPd5qPzvTm5hUt6X23TJdh9H4NjNsNbt7t7UuTMJfawQWsdWRFhfLwkiMkB1rQ4ZJWLB9YBnzR7kbs9N8b2PsKZgKUHQm1X4or",
	ChainParams: &pb.ChainParams{
		Network: &pb.ChainParams_BitcoinNetwork{
			BitcoinNetwork: pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET,
		},
	},
	Scheme: pb.Scheme_SCHEME_BIP44,
}

var BitcoinTestnet3P2PKH = Fixture{
	ExternalDescriptor: "pkh(tpubDC5FSnBiZDMmhiuCmWAYsLwgLYrrT9rAqvTySfuCCrgsWz8wxMXUS9Tb9iVMvcRbvFcAHGkMD5Kx8koh4GquNGNTfohfk7pgjhaPCdXpoba/0/*)",
	InternalDescriptor: "pkh(tpubDC5FSnBiZDMmhiuCmWAYsLwgLYrrT9rAqvTySfuCCrgsWz8wxMXUS9Tb9iVMvcRbvFcAHGkMD5Kx8koh4GquNGNTfohfk7pgjhaPCdXpoba/1/*)",
	ExtendedPublicKey:  "tpubDC5FSnBiZDMmhiuCmWAYsLwgLYrrT9rAqvTySfuCCrgsWz8wxMXUS9Tb9iVMvcRbvFcAHGkMD5Kx8koh4GquNGNTfohfk7pgjhaPCdXpoba",
	ChainParams: &pb.ChainParams{
		Network: &pb.ChainParams_BitcoinNetwork{
			BitcoinNetwork: pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3,
		},
	},
	Scheme: pb.Scheme_SCHEME_BIP44,
}

var BitcoinTestnet3P2SHP2WPKH = Fixture{
	ExternalDescriptor: "sh(wpkh(tpubDCcvqEHx7prGddpWTfEviiew5YLMrrKy4oJbt14teJZenSi6AYMAs2SNXwYXFzkrNYwECSmobwxESxMCrpfqw4gsUt88bcr8iMrJmbb8P2q/0/*))",
	InternalDescriptor: "sh(wpkh(tpubDCcvqEHx7prGddpWTfEviiew5YLMrrKy4oJbt14teJZenSi6AYMAs2SNXwYXFzkrNYwECSmobwxESxMCrpfqw4gsUt88bcr8iMrJmbb8P2q/1/*))",
	ExtendedPublicKey:  "tpubDCcvqEHx7prGddpWTfEviiew5YLMrrKy4oJbt14teJZenSi6AYMAs2SNXwYXFzkrNYwECSmobwxESxMCrpfqw4gsUt88bcr8iMrJmbb8P2q",
	ChainParams: &pb.ChainParams{
		Network: &pb.ChainParams_BitcoinNetwork{
			BitcoinNetwork: pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3,
		},
	},
	Scheme: pb.Scheme_SCHEME_BIP49,
}

var BitcoinMainnetP2WPKH = Fixture{
	ExternalDescriptor: "wpkh(xpub6CMeLkY9TzXyLYXPWMXB5LWtprVABb6HwPEPXnEgESMNrSUBsvhXNsA7zKS1ZRKhUyQG4HjZysEP8v7gDNU4J6PvN5yLx4meEm3mpEapLMN/0/*)",
	InternalDescriptor: "wpkh(xpub6CMeLkY9TzXyLYXPWMXB5LWtprVABb6HwPEPXnEgESMNrSUBsvhXNsA7zKS1ZRKhUyQG4HjZysEP8v7gDNU4J6PvN5yLx4meEm3mpEapLMN/1/*)",
	ExtendedPublicKey:  "xpub6CMeLkY9TzXyLYXPWMXB5LWtprVABb6HwPEPXnEgESMNrSUBsvhXNsA7zKS1ZRKhUyQG4HjZysEP8v7gDNU4J6PvN5yLx4meEm3mpEapLMN",
	ChainParams: &pb.ChainParams{
		Network: &pb.ChainParams_BitcoinNetwork{
			BitcoinNetwork: pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET,
		},
	},
	Scheme: pb.Scheme_SCHEME_BIP84,
}

var LitecoinMainnetP2WPKH = Fixture{
	ExternalDescriptor: "wpkh(Ltub2YC8XgcRjMJqvX8LsuBxdM7PKE5uih6247CpgK2rfEdzEGt1YHVHW4L865ss5eEy2K1KixTMkrHJbzTtqxpiGpM4wyrxYRFJFxuACSJqkyo/0/*)",
	InternalDescriptor: "wpkh(Ltub2YC8XgcRjMJqvX8LsuBxdM7PKE5uih6247CpgK2rfEdzEGt1YHVHW4L865ss5eEy2K1KixTMkrHJbzTtqxpiGpM4wyrxYRFJFxuACSJqkyo/1/*)",
	ExtendedPublicKey:  "Ltub2YC8XgcRjMJqvX8LsuBxdM7PKE5uih6247CpgK2rfEdzEGt1YHVHW4L865ss5eEy2K1KixTMkrHJbzTtqxpiGpM4wyrxYRFJFxuACSJqkyo",
	ChainParams: &pb.ChainParams{
		Network: &pb.ChainParams_LitecoinNetwork{
			LitecoinNetwork: pb.LitecoinNetwork_LITECOIN_NETWORK_MAINNET,
		},
	},
	Scheme: pb.Scheme_SCHEME_BIP84,
}
