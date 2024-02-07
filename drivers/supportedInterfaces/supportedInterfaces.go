package supportedinterfaces

const (
	V1_API    = "v1"
	V2_API    = "v2"
	GRPC_CALL = "grpc"
)

var (
	API_V1_SUPPORTED_METHODS = [...]string{
		"GetAccountList",
		"GetAccount",
	}

	API_V2_SUPPORTED_METHODS = [...]string{
		"GetAccountList",
		"GetAccount",
	}

	GRPC_SUPPORTED_METHODS = [...]string{
		"GetAccountList",
		"GetAccount",
	}

	SupportedMethods = map[string][]string{
		V1_API:    API_V1_SUPPORTED_METHODS[:],
		V2_API:    API_V2_SUPPORTED_METHODS[:],
		GRPC_CALL: GRPC_SUPPORTED_METHODS[:],
	}
)
