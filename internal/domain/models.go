package domain

// LocalizedText keeps the API explicit for clients that do not perform locale negotiation.
type LocalizedText struct {
	ZH string `json:"zh"`
	EN string `json:"en"`
	ES string `json:"es,omitempty"`
	JA string `json:"ja,omitempty"`
	FR string `json:"fr,omitempty"`
	KO string `json:"ko,omitempty"`
}

type Authorization struct {
	Holder          LocalizedText `json:"holder"`
	Scope           LocalizedText `json:"scope"`
	Territories     []string      `json:"territories"`
	UsageTypes      []string      `json:"usageTypes"`
	Exclusivity     string        `json:"exclusivity"`
	CommercialUse   bool          `json:"commercialUse"`
	DerivativeWorks bool          `json:"derivativeWorks"`
	AITraining      bool          `json:"aiTraining"`
	Sublicensable   bool          `json:"sublicensable"`
	ValidFrom       string        `json:"validFrom"`
	ValidUntil      string        `json:"validUntil,omitempty"`
}

type Provenance struct {
	Issuer             string `json:"issuer"`
	CertificateID      string `json:"certificateId"`
	Network            string `json:"network"`
	TokenStandard      string `json:"tokenStandard"`
	ContractAddress    string `json:"contractAddress,omitempty"`
	MetadataURI        string `json:"metadataUri,omitempty"`
	VerificationStatus string `json:"verificationStatus"`
	IssuedAt           string `json:"issuedAt"`
}

type AssetMedia struct {
	LinkedWorkIDs []string `json:"linkedWorkIds"`
	Format        string   `json:"format"`
	PreviewURL    string   `json:"previewUrl,omitempty"`
}

// ExternalAssetRef identifies the upstream platform record used to build the
// normalized HAPW view. Clipli does not become the source of truth for assets.
type ExternalAssetRef struct {
	ProviderCode    string `json:"providerCode"`
	ProviderAssetID string `json:"providerAssetId"`
	AssetURL        string `json:"assetUrl,omitempty"`
	SyncStatus      string `json:"syncStatus"`
	LastSyncedAt    string `json:"lastSyncedAt"`
	DataVersion     string `json:"dataVersion,omitempty"`
}

// HAPWAsset retains the original flat fields for the current frontend and adds
// structured authorization and provenance blocks for API consumers.
type HAPWAsset struct {
	ID                   string           `json:"id"`
	Kind                 string           `json:"kind"`
	TokenID              string           `json:"tokenId"`
	Name                 string           `json:"name"`
	NameEn               string           `json:"nameEn"`
	NameKo               string           `json:"nameKo,omitempty"`
	Value                int              `json:"value"`
	Currency             string           `json:"currency"`
	Status               string           `json:"status"`
	StatusEn             string           `json:"statusEn"`
	StatusKo             string           `json:"statusKo,omitempty"`
	Transferable         bool             `json:"transferable"`
	Owner                string           `json:"owner"`
	RightsHolder         string           `json:"rightsHolder"`
	RightsHolderEn       string           `json:"rightsHolderEn"`
	RightsHolderKo       string           `json:"rightsHolderKo,omitempty"`
	AuthorizationScope   string           `json:"authorizationScope"`
	AuthorizationScopeEn string           `json:"authorizationScopeEn"`
	AuthorizationScopeKo string           `json:"authorizationScopeKo,omitempty"`
	CreditYield          int              `json:"creditYield"`
	ClipPrice            int              `json:"clipPrice"`
	ExchangeAvailable    bool             `json:"exchangeAvailable"`
	RedemptionStatus     string           `json:"redemptionStatus"`
	Authorization        Authorization    `json:"authorization"`
	Provenance           Provenance       `json:"provenance"`
	Media                AssetMedia       `json:"media"`
	External             ExternalAssetRef `json:"external"`
}

type Work struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	TitleEn       string `json:"titleEn"`
	TitleKo       string `json:"titleKo,omitempty"`
	Category      string `json:"category"`
	CategoryEn    string `json:"categoryEn"`
	CategoryKo    string `json:"categoryKo,omitempty"`
	Duration      string `json:"duration"`
	Views         int    `json:"views"`
	Creator       string `json:"creator"`
	CreatorEn     string `json:"creatorEn"`
	CreatorKo     string `json:"creatorKo,omitempty"`
	Summary       string `json:"summary"`
	SummaryEn     string `json:"summaryEn"`
	SummaryKo     string `json:"summaryKo,omitempty"`
	Description   string `json:"description"`
	DescriptionEn string `json:"descriptionEn"`
	DescriptionKo string `json:"descriptionKo,omitempty"`
	Format        string `json:"format"`
	FormatEn      string `json:"formatEn"`
	FormatKo      string `json:"formatKo,omitempty"`
	LinkedAssetID string `json:"linkedAssetId"`
	HaiwenURL     string `json:"haiwenUrl"`
	Accent        string `json:"accent"`
	Source        string `json:"source"`
}

type AssetExercise struct {
	ID           string `json:"id"`
	RequestID    string `json:"requestId"`
	AssetID      string `json:"assetId"`
	PlatformCode string `json:"platformCode"`
	Direction    string `json:"direction"`
	DirectionEn  string `json:"directionEn"`
	DirectionKo  string `json:"directionKo,omitempty"`
	Value        int    `json:"value"`
	StatusCode   string `json:"statusCode"`
	Status       string `json:"status"`
	StatusEn     string `json:"statusEn"`
	StatusKo     string `json:"statusKo,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type CLIPPool struct {
	ClipReserve int    `json:"clipReserve"`
	USDTReserve int    `json:"usdtReserve"`
	UpdatedAt   string `json:"updatedAt"`
}

type CLIPAccount struct {
	Symbol              string   `json:"symbol"`
	Balance             int      `json:"balance"`
	SupplyPolicy        string   `json:"supplyPolicy"`
	SupplyPolicyEn      string   `json:"supplyPolicyEn"`
	Acquisition         string   `json:"acquisition"`
	AcquisitionEn       string   `json:"acquisitionEn"`
	AcquisitionKo       string   `json:"acquisitionKo,omitempty"`
	QuoteAsset          string   `json:"quoteAsset"`
	PricePolicy         string   `json:"pricePolicy"`
	PricePolicyEn       string   `json:"pricePolicyEn"`
	PricePolicyKo       string   `json:"pricePolicyKo,omitempty"`
	ContractStatus      string   `json:"contractStatus"`
	ContractStatusEn    string   `json:"contractStatusEn"`
	ContractStatusKo    string   `json:"contractStatusKo,omitempty"`
	DexURL              string   `json:"dexUrl"`
	DexPool             CLIPPool `json:"dexPool"`
	HAPWExchangeFeeRate float64  `json:"hapwExchangeFeeRate"`
}

type CLIPTransaction struct {
	ID             string `json:"id"`
	TypeCode       string `json:"typeCode"`
	Type           string `json:"type"`
	TypeEn         string `json:"typeEn"`
	TypeKo         string `json:"typeKo,omitempty"`
	Amount         int    `json:"amount"`
	Counterparty   string `json:"counterparty"`
	CounterpartyEn string `json:"counterpartyEn"`
	CounterpartyKo string `json:"counterpartyKo,omitempty"`
	StatusCode     string `json:"statusCode"`
	Status         string `json:"status"`
	StatusEn       string `json:"statusEn"`
	StatusKo       string `json:"statusKo,omitempty"`
	TxHash         string `json:"txHash"`
	CreatedAt      string `json:"createdAt"`
}

// CLIPTreasury is the conservation ledger for the fixed, one-time mint. The
// public contract never exposes a mint operation; distribution is performed by
// the platform service against centralized rules.
type CLIPTreasury struct {
	Symbol              string `json:"symbol"`
	MintMode            string `json:"mintMode"`
	MintedSupply        int    `json:"mintedSupply"`
	TreasuryBalance     int    `json:"treasuryBalance"`
	LiquidityAllocation int    `json:"liquidityAllocation"`
	LedgerOutstanding   int    `json:"ledgerOutstanding"`
	TotalDistributed    int    `json:"totalDistributed"`
	TotalReclaimed      int    `json:"totalReclaimed"`
	PlatformWallet      string `json:"platformWallet"`
	ContractAddress     string `json:"contractAddress"`
	Network             string `json:"network"`
	DistributionMode    string `json:"distributionMode"`
	MintStatus          string `json:"mintStatus"`
	MintedAt            string `json:"mintedAt"`
}

type CLIPDistributionRule struct {
	Code        string  `json:"code"`
	Trigger     string  `json:"trigger"`
	Formula     string  `json:"formula"`
	Rate        float64 `json:"rate"`
	Description string  `json:"description"`
	Enabled     bool    `json:"enabled"`
}

type CLIPDistribution struct {
	ID                string `json:"id"`
	RequestID         string `json:"requestId"`
	RuleCode          string `json:"ruleCode"`
	UserRef           string `json:"userRef"`
	AssetID           string `json:"assetId"`
	Amount            int    `json:"amount"`
	UserBalanceBefore int    `json:"userBalanceBefore"`
	UserBalanceAfter  int    `json:"userBalanceAfter"`
	TreasuryBefore    int    `json:"treasuryBefore"`
	TreasuryAfter     int    `json:"treasuryAfter"`
	Status            string `json:"status"`
	CreatedAt         string `json:"createdAt"`
}

type AssetSource struct {
	Code                string `json:"code"`
	Name                string `json:"name"`
	BaseURL             string `json:"baseUrl"`
	Mode                string `json:"mode"`
	Status              string `json:"status"`
	SyncIntervalSeconds int    `json:"syncIntervalSeconds"`
	LastSyncedAt        string `json:"lastSyncedAt,omitempty"`
	LastError           string `json:"lastError,omitempty"`
	AssetCount          int    `json:"assetCount"`
}

type AssetSyncRun struct {
	ID           string `json:"id"`
	SourceCode   string `json:"sourceCode"`
	Status       string `json:"status"`
	StartedAt    string `json:"startedAt"`
	CompletedAt  string `json:"completedAt,omitempty"`
	RecordsRead  int    `json:"recordsRead"`
	RecordsValid int    `json:"recordsValid"`
	RecordsSaved int    `json:"recordsSaved"`
	Error        string `json:"error,omitempty"`
}

type SessionPolicy struct {
	Mode                 string `json:"mode"`
	UserRef              string `json:"userRef"`
	IdentityVerification bool   `json:"identityVerification"`
	KYCRequired          bool   `json:"kycRequired"`
	WalletOptional       bool   `json:"walletOptional"`
	Persistence          string `json:"persistence"`
	Notice               string `json:"notice"`
}

type GenerationAccount struct {
	Balance             int     `json:"balance"`
	LifetimeGranted     int     `json:"lifetimeGranted"`
	LifetimeUsed        int     `json:"lifetimeUsed"`
	StandardFactor      float64 `json:"standardFactor"`
	ProFactor           float64 `json:"proFactor"`
	ClipGrantPerCredit  float64 `json:"clipGrantPerCredit"`
	ClipCostPerCredit   float64 `json:"clipCostPerCredit"`
	LifetimeClipGranted int     `json:"lifetimeClipGranted"`
	LifetimeClipSpent   int     `json:"lifetimeClipSpent"`
}

type HAPWRedemption struct {
	ID               string `json:"id"`
	AssetID          string `json:"assetId"`
	Receipt          string `json:"receipt"`
	CreditsGranted   int    `json:"creditsGranted"`
	CreditsRemaining int    `json:"creditsRemaining"`
	ClipGranted      int    `json:"clipGranted"`
	RequestID        string `json:"requestId"`
	Status           string `json:"status"`
	StatusEn         string `json:"statusEn"`
	StatusKo         string `json:"statusKo,omitempty"`
	CreatedAt        string `json:"createdAt"`
}

type Generation struct {
	ID          string `json:"id"`
	AssetID     string `json:"assetId"`
	Title       string `json:"title"`
	TitleEn     string `json:"titleEn"`
	TitleKo     string `json:"titleKo,omitempty"`
	Duration    int    `json:"duration"`
	Quality     string `json:"quality"`
	CreditsUsed int    `json:"creditsUsed"`
	ValidViews  int    `json:"validViews"`
	ClipCost    int    `json:"clipCost"`
	RequestID   string `json:"requestId"`
	Status      string `json:"status"`
	StatusEn    string `json:"statusEn"`
	StatusKo    string `json:"statusKo,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

type Conversion struct {
	ID         string `json:"id"`
	RequestID  string `json:"requestId"`
	AssetID    string `json:"assetId"`
	Asset      string `json:"asset"`
	Region     string `json:"region"`
	Days       int    `json:"days"`
	Fee        int    `json:"fee"`
	StatusCode string `json:"statusCode"`
	Status     string `json:"status"`
	StatusEn   string `json:"statusEn"`
	StatusKo   string `json:"statusKo,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

type HAPWExchange struct {
	ID         string  `json:"id"`
	RequestID  string  `json:"requestId"`
	AssetID    string  `json:"assetId"`
	Price      int     `json:"price"`
	Fee        int     `json:"fee"`
	Total      int     `json:"total"`
	FeeRate    float64 `json:"feeRate"`
	StatusCode string  `json:"statusCode"`
	Status     string  `json:"status"`
	StatusEn   string  `json:"statusEn"`
	StatusKo   string  `json:"statusKo,omitempty"`
	CreatedAt  string  `json:"createdAt"`
}

type ExternalPlatform struct {
	Code string `json:"code"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ProfileSettings struct {
	WalletSign     bool `json:"walletSign"`
	ExpiryReminder bool `json:"expiryReminder"`
}

type Profile struct {
	Wallet          string          `json:"wallet"`
	WalletProvider  string          `json:"walletProvider"`
	OverseasAccount string          `json:"overseasAccount"`
	Phone           string          `json:"phone"`
	Level           int             `json:"level"`
	Points          int             `json:"points"`
	Settings        ProfileSettings `json:"settings"`
}

type AssetEvent struct {
	ID        string         `json:"id"`
	AssetID   string         `json:"assetId"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	CreatedAt string         `json:"createdAt"`
	Details   map[string]any `json:"details,omitempty"`
}
