package adaptor

type adapterCryptoCurrency interface {
	NewPrivateKey() []byte
	GetPublicKey(key []byte) (pubkey []byte)
	GetAddress(key []byte) (address string)
	CreateMultiSigAddress(params string)
	GetUnspendUTXO(params string) string
	RawTransactionGen(params string) string
	DecodeRawTransaction(params string) string
	SignTransaction(params string) string
	GetBalance(params string) string
	GetTransactions(params string) string
}

type adapterSmartContract interface {
}
