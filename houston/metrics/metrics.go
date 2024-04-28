package metrics

type TransactionMetricsCollector interface {
	// number of transactions
	getTransactionCount() (int, error)
	// number of different message types
	getMessageTypesCount() (int, error)
	// most active transaction senders
	getTopTransactionSenders() ([]string, error)
	// most active realms deployed
	getMostActiveRealmsDeployed() ([]string, error)
	// most active realms called
	getMostActiveRealmsCalled() ([]string, error)
	// most active packages deployed
	getMostActivePackagesDeployed() ([]string, error)
	// most active packages called
	getMostActivePackagesCalled() ([]string, error)
}
