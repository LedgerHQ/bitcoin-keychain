package integration

import pb "github.com/ledgerhq/bitcoin-keychain-svc/pb/keychain"

type Fixture struct {
	Descriptor string
	XPub       string
	Network    pb.BitcoinNetwork
	Scheme     pb.KeychainInfo_Scheme
}

var BitcoinMainnetP2PKH = Fixture{
	Descriptor: "pkh(xpub6DCi5iJ57ZPd5qPzvTm5hUt6X23TJdh9H4NjNsNbt7t7UuTMJfawQWsdWRFhfLwkiMkB1rQ4ZJWLB9YBnzR7kbs9N8b2PsKZgKUHQm1X4or)",
	XPub:       "xpub6DCi5iJ57ZPd5qPzvTm5hUt6X23TJdh9H4NjNsNbt7t7UuTMJfawQWsdWRFhfLwkiMkB1rQ4ZJWLB9YBnzR7kbs9N8b2PsKZgKUHQm1X4or",
	Network:    pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET,
	Scheme:     pb.KeychainInfo_SCHEME_BIP44,
}

var BitcoinTestnet3P2PKH = Fixture{
	Descriptor: "pkh(tpubDC5FSnBiZDMmhiuCmWAYsLwgLYrrT9rAqvTySfuCCrgsWz8wxMXUS9Tb9iVMvcRbvFcAHGkMD5Kx8koh4GquNGNTfohfk7pgjhaPCdXpoba)",
	XPub:       "tpubDC5FSnBiZDMmhiuCmWAYsLwgLYrrT9rAqvTySfuCCrgsWz8wxMXUS9Tb9iVMvcRbvFcAHGkMD5Kx8koh4GquNGNTfohfk7pgjhaPCdXpoba",
	Network:    pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3,
	Scheme:     pb.KeychainInfo_SCHEME_BIP44,
}

var BitcoinTestnet3P2SHP2WPKH = Fixture{
	Descriptor: "sh(wpkh(tpubDCcvqEHx7prGddpWTfEviiew5YLMrrKy4oJbt14teJZenSi6AYMAs2SNXwYXFzkrNYwECSmobwxESxMCrpfqw4gsUt88bcr8iMrJmbb8P2q))",
	XPub:       "tpubDCcvqEHx7prGddpWTfEviiew5YLMrrKy4oJbt14teJZenSi6AYMAs2SNXwYXFzkrNYwECSmobwxESxMCrpfqw4gsUt88bcr8iMrJmbb8P2q",
	Network:    pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3,
	Scheme:     pb.KeychainInfo_SCHEME_BIP49,
}

var BitcoinMainnetP2WPKH = Fixture{
	Descriptor: "wpkh(xpub6CMeLkY9TzXyLYXPWMXB5LWtprVABb6HwPEPXnEgESMNrSUBsvhXNsA7zKS1ZRKhUyQG4HjZysEP8v7gDNU4J6PvN5yLx4meEm3mpEapLMN)",
	XPub:       "xpub6CMeLkY9TzXyLYXPWMXB5LWtprVABb6HwPEPXnEgESMNrSUBsvhXNsA7zKS1ZRKhUyQG4HjZysEP8v7gDNU4J6PvN5yLx4meEm3mpEapLMN",
	Network:    pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET,
	Scheme:     pb.KeychainInfo_SCHEME_BIP84,
}
