package adaptor

type adapterCryptoCurrency interface {
	NewPrivateKey() (wifPriKey string)
	GetPublicKey(wifPriKey string) (pubKey string)
	GetAddress(wifPriKey string) (address string)
	CreateMultiSigAddress(params string) string
	GetUnspendUTXO(params string) string
	RawTransactionGen(params string) string
	DecodeRawTransaction(params string) string
	SignTransaction(params string) string
	GetBalance(params string) string
	GetTransactions(params string) string
	ImportMultisig(params string) string
}

type adapterSmartContract interface {
}
