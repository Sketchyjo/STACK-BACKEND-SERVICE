package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Environment  string             `mapstructure:"environment"`
	LogLevel     string             `mapstructure:"log_level"`
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	Blockchain   BlockchainConfig   `mapstructure:"blockchain"`
	Payment      PaymentConfig      `mapstructure:"payment"`
	Security     SecurityConfig     `mapstructure:"security"`
	Circle       CircleConfig       `mapstructure:"circle"`
	KYC          KYCConfig          `mapstructure:"kyc"`
	Email        EmailConfig        `mapstructure:"email"`
	SMS          SMSConfig          `mapstructure:"sms"`
	Verification VerificationConfig `mapstructure:"verification"`
	ZeroG        ZeroGConfig        `mapstructure:"zerog"`
}

type ServerConfig struct {
	Port            int      `mapstructure:"port"`
	Host            string   `mapstructure:"host"`
	ReadTimeout     int      `mapstructure:"read_timeout"`
	WriteTimeout    int      `mapstructure:"write_timeout"`
	AllowedOrigins  []string `mapstructure:"allowed_origins"`
	RateLimitPerMin int      `mapstructure:"rate_limit_per_min"`
}

type DatabaseConfig struct {
	URL             string `mapstructure:"url"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	AccessTTL  int    `mapstructure:"access_token_ttl"`
	RefreshTTL int    `mapstructure:"refresh_token_ttl"`
	Issuer     string `mapstructure:"issuer"`
}

type BlockchainConfig struct {
	Networks map[string]NetworkConfig `mapstructure:"networks"`
}

type NetworkConfig struct {
	Name           string                 `mapstructure:"name"`
	ChainID        int                    `mapstructure:"chain_id"`
	RPC            string                 `mapstructure:"rpc"`
	WebSocket      string                 `mapstructure:"websocket"`
	Explorer       string                 `mapstructure:"explorer"`
	NativeCurrency CurrencyConfig         `mapstructure:"native_currency"`
	Tokens         map[string]TokenConfig `mapstructure:"tokens"`
	GasLimit       int                    `mapstructure:"gas_limit"`
	MaxGasPrice    string                 `mapstructure:"max_gas_price"`
}

type CurrencyConfig struct {
	Name     string `mapstructure:"name"`
	Symbol   string `mapstructure:"symbol"`
	Decimals int    `mapstructure:"decimals"`
}

type TokenConfig struct {
	Address  string `mapstructure:"address"`
	Symbol   string `mapstructure:"symbol"`
	Name     string `mapstructure:"name"`
	Decimals int    `mapstructure:"decimals"`
	ChainID  int    `mapstructure:"chain_id"`
}

type PaymentConfig struct {
	ProcessorAPIKey string              `mapstructure:"processor_api_key"`
	WebhookSecret   string              `mapstructure:"webhook_secret"`
	Cards           CardProcessorConfig `mapstructure:"cards"`
	Supported       []string            `mapstructure:"supported_currencies"`
}

type CardProcessorConfig struct {
	Provider    string `mapstructure:"provider"`
	APIKey      string `mapstructure:"api_key"`
	APISecret   string `mapstructure:"api_secret"`
	WebhookURL  string `mapstructure:"webhook_url"`
	Environment string `mapstructure:"environment"` // sandbox, production
}

type SecurityConfig struct {
	EncryptionKey     string   `mapstructure:"encryption_key"`
	AllowedIPs        []string `mapstructure:"allowed_ips"`
	MaxLoginAttempts  int      `mapstructure:"max_login_attempts"`
	LockoutDuration   int      `mapstructure:"lockout_duration"`
	RequireMFA        bool     `mapstructure:"require_mfa"`
	PasswordMinLength int      `mapstructure:"password_min_length"`
	SessionTimeout    int      `mapstructure:"session_timeout"`
}

type CircleConfig struct {
	APIKey                 string   `mapstructure:"api_key"`
	Environment            string   `mapstructure:"environment"` // sandbox or production
	BaseURL                string   `mapstructure:"base_url"`
	EntitySecretCiphertext string   `mapstructure:"entity_secret_ciphertext"`
	DefaultWalletSetID     string   `mapstructure:"default_wallet_set_id"`
	DefaultWalletSetName   string   `mapstructure:"default_wallet_set_name"`
	SupportedChains        []string `mapstructure:"supported_chains"`
}

type KYCConfig struct {
	Provider    string `mapstructure:"provider"` // "sumsub", "jumio"
	APIKey      string `mapstructure:"api_key"`
	APISecret   string `mapstructure:"api_secret"`
	BaseURL     string `mapstructure:"base_url"`
	CallbackURL string `mapstructure:"callback_url"`
	Environment string `mapstructure:"environment"` // "development", "sandbox", "production"
	UserAgent   string `mapstructure:"user_agent"`
	LevelName   string `mapstructure:"level_name"`
}

type EmailConfig struct {
	Provider    string `mapstructure:"provider"` // "sendgrid", "resend"
	APIKey      string `mapstructure:"api_key"`
	FromEmail   string `mapstructure:"from_email"`
	FromName    string `mapstructure:"from_name"`
	BaseURL     string `mapstructure:"base_url"`    // For verification links
	Environment string `mapstructure:"environment"` // "development", "staging", "production"
	ReplyTo     string `mapstructure:"reply_to"`
}

type SMSConfig struct {
	Provider    string `mapstructure:"provider"` // "twilio"
	APIKey      string `mapstructure:"api_key"`
	APISecret   string `mapstructure:"api_secret"`
	FromNumber  string `mapstructure:"from_number"`
	Environment string `mapstructure:"environment"` // "development", "staging", "production"
}

type VerificationConfig struct {
	CodeLength       int `mapstructure:"code_length"`
	CodeTTLMinutes   int `mapstructure:"code_ttl_minutes"`
	MaxAttempts      int `mapstructure:"max_attempts"`
	RateLimitPerHour int `mapstructure:"rate_limit_per_hour"`
}

// ZeroGConfig contains configuration for 0G Network integration
type ZeroGConfig struct {
	// Storage configuration
	Storage ZeroGStorageConfig `mapstructure:"storage"`
	// Compute/Inference configuration
	Compute ZeroGComputeConfig `mapstructure:"compute"`
	// General settings
	Timeout        int  `mapstructure:"timeout"`          // Request timeout in seconds
	MaxRetries     int  `mapstructure:"max_retries"`      // Maximum retry attempts
	RetryBackoffMs int  `mapstructure:"retry_backoff_ms"` // Retry backoff in milliseconds
	EnableMetrics  bool `mapstructure:"enable_metrics"`   // Enable observability metrics
	EnableTracing  bool `mapstructure:"enable_tracing"`   // Enable distributed tracing
}

// ZeroGStorageConfig contains 0G storage specific configuration
type ZeroGStorageConfig struct {
	RPCEndpoint      string          `mapstructure:"rpc_endpoint"`      // 0G storage RPC endpoint
	IndexerRPC       string          `mapstructure:"indexer_rpc"`       // 0G indexer RPC endpoint
	PrivateKey       string          `mapstructure:"private_key"`       // Private key for storage operations
	MinReplicas      int             `mapstructure:"min_replicas"`      // Minimum replication count
	ExpectedReplicas int             `mapstructure:"expected_replicas"` // Expected replication count
	Namespaces       ZeroGNamespaces `mapstructure:"namespaces"`        // Storage namespaces
}

// ZeroGComputeConfig contains 0G compute/inference specific configuration
type ZeroGComputeConfig struct {
	BrokerEndpoint string           `mapstructure:"broker_endpoint"` // 0G compute broker endpoint
	PrivateKey     string           `mapstructure:"private_key"`     // Private key for compute operations
	ProviderID     string           `mapstructure:"provider_id"`     // Preferred inference provider ID
	ModelConfig    ZeroGModelConfig `mapstructure:"model_config"`    // AI model configuration
	Funding        ZeroGFunding     `mapstructure:"funding"`         // Account funding configuration
}

// ZeroGNamespaces contains predefined storage namespaces
type ZeroGNamespaces struct {
	AISummaries  string `mapstructure:"ai_summaries"`  // ai-summaries/ namespace
	AIArtifacts  string `mapstructure:"ai_artifacts"`  // ai-artifacts/ namespace
	ModelPrompts string `mapstructure:"model_prompts"` // model-prompts/ namespace
}

// ZeroGModelConfig contains AI model configuration
type ZeroGModelConfig struct {
	DefaultModel     string  `mapstructure:"default_model"`     // Default LLM model to use
	MaxTokens        int     `mapstructure:"max_tokens"`        // Maximum tokens per request
	Temperature      float64 `mapstructure:"temperature"`       // Model temperature setting
	TopP             float64 `mapstructure:"top_p"`             // Top-p sampling parameter
	FrequencyPenalty float64 `mapstructure:"frequency_penalty"` // Frequency penalty
	PresencePenalty  float64 `mapstructure:"presence_penalty"`  // Presence penalty
}

// ZeroGFunding contains account funding configuration
type ZeroGFunding struct {
	AutoTopup       bool    `mapstructure:"auto_topup"`        // Enable automatic balance top-up
	MinBalance      float64 `mapstructure:"min_balance"`       // Minimum account balance threshold
	TopupAmount     float64 `mapstructure:"topup_amount"`      // Amount to top up when threshold reached
	MaxAccountLimit float64 `mapstructure:"max_account_limit"` // Maximum account balance limit
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	// Load .env file if it exists (ignore errors if file doesn't exist)
	godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Read from config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Override specific environment variables
	overrideFromEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Build database URL if not provided
	if config.Database.URL == "" {
		config.Database.URL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Name,
			config.Database.SSLMode,
		)
	}

	// Validate required fields
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.rate_limit_per_min", 100)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "stack_service")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", 300)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	// JWT defaults
	viper.SetDefault("jwt.access_token_ttl", 3600)     // 1 hour
	viper.SetDefault("jwt.refresh_token_ttl", 2592000) // 30 days
	viper.SetDefault("jwt.issuer", "stack_service")

	// Security defaults
	viper.SetDefault("security.max_login_attempts", 5)
	viper.SetDefault("security.lockout_duration", 900) // 15 minutes
	viper.SetDefault("security.require_mfa", false)
	viper.SetDefault("security.password_min_length", 8)

	// Circle defaults
	viper.SetDefault("circle.environment", "sandbox")
	viper.SetDefault("circle.api_key", "")
	viper.SetDefault("circle.base_url", "")
	viper.SetDefault("circle.entity_secret_ciphertext", "")
	viper.SetDefault("circle.default_wallet_set_id", "")
	viper.SetDefault("circle.default_wallet_set_name", "STACK-WalletSet")
	viper.SetDefault("circle.supported_chains", []string{"ETH", "MATIC", "SOL", "BASE"})

	// KYC defaults
	viper.SetDefault("kyc.provider", "")
	viper.SetDefault("kyc.environment", "development")
	viper.SetDefault("kyc.base_url", "https://netverify.com")
	viper.SetDefault("kyc.user_agent", "Stack-Service/1.0")
	viper.SetDefault("kyc.level_name", "basic-kyc")

	// Email defaults
	viper.SetDefault("email.provider", "")
	viper.SetDefault("email.from_email", "no-reply@stackservice.com")
	viper.SetDefault("email.from_name", "Stack Service")
	viper.SetDefault("email.environment", "development")
	viper.SetDefault("email.base_url", "http://localhost:3000")
	viper.SetDefault("email.reply_to", "")

	// SMS defaults
	viper.SetDefault("sms.provider", "")
	viper.SetDefault("sms.environment", "development")

	// Verification defaults
	viper.SetDefault("verification.code_length", 6)
	viper.SetDefault("verification.code_ttl_minutes", 10)
	viper.SetDefault("verification.max_attempts", 3)
	viper.SetDefault("verification.rate_limit_per_hour", 3)

	viper.SetDefault("security.session_timeout", 3600) // 1 hour

	// 0G Network defaults
	// General 0G settings
	viper.SetDefault("zerog.timeout", 30)            // 30 seconds
	viper.SetDefault("zerog.max_retries", 3)         // 3 retry attempts
	viper.SetDefault("zerog.retry_backoff_ms", 1000) // 1 second
	viper.SetDefault("zerog.enable_metrics", true)   // Enable metrics
	viper.SetDefault("zerog.enable_tracing", true)   // Enable tracing

	// Storage defaults - 0G Testnet endpoints
	viper.SetDefault("zerog.storage.rpc_endpoint", "https://evmrpc-testnet.0g.ai/")
	viper.SetDefault("zerog.storage.indexer_rpc", "https://indexer-storage-testnet-turbo.0g.ai")
	viper.SetDefault("zerog.storage.min_replicas", 1)
	viper.SetDefault("zerog.storage.expected_replicas", 3)
	viper.SetDefault("zerog.storage.namespaces.ai_summaries", "ai-summaries/")
	viper.SetDefault("zerog.storage.namespaces.ai_artifacts", "ai-artifacts/")
	viper.SetDefault("zerog.storage.namespaces.model_prompts", "model-prompts/")

	// Compute defaults
	viper.SetDefault("zerog.compute.broker_endpoint", "")
	viper.SetDefault("zerog.compute.provider_id", "")
	viper.SetDefault("zerog.compute.model_config.default_model", "gpt-4")
	viper.SetDefault("zerog.compute.model_config.max_tokens", 4096)
	viper.SetDefault("zerog.compute.model_config.temperature", 0.7)
	viper.SetDefault("zerog.compute.model_config.top_p", 0.9)
	viper.SetDefault("zerog.compute.model_config.frequency_penalty", 0.0)
	viper.SetDefault("zerog.compute.model_config.presence_penalty", 0.0)
	viper.SetDefault("zerog.compute.funding.auto_topup", false)
	viper.SetDefault("zerog.compute.funding.min_balance", 10.0)
	viper.SetDefault("zerog.compute.funding.topup_amount", 50.0)
	viper.SetDefault("zerog.compute.funding.max_account_limit", 1000.0)
}

func overrideFromEnv() {
	// Server
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			viper.Set("server.port", p)
		}
	}

	// Database
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		viper.Set("database.url", dbURL)
	}

	// JWT
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		viper.Set("jwt.secret", jwtSecret)
	}

	// Encryption
	if encKey := os.Getenv("ENCRYPTION_KEY"); encKey != "" {
		viper.Set("security.encryption_key", encKey)
	}

	// Circle API
	if circleKey := os.Getenv("CIRCLE_API_KEY"); circleKey != "" {
		viper.Set("circle.api_key", circleKey)
	}
	if circleBaseURL := os.Getenv("CIRCLE_BASE_URL"); circleBaseURL != "" {
		viper.Set("circle.base_url", circleBaseURL)
	}
	if circleEntitySecret := os.Getenv("CIRCLE_ENTITY_SECRET_CIPHERTEXT"); circleEntitySecret != "" {
		viper.Set("circle.entity_secret_ciphertext", circleEntitySecret)
	}
	if circleWalletSetID := os.Getenv("CIRCLE_DEFAULT_WALLET_SET_ID"); circleWalletSetID != "" {
		viper.Set("circle.default_wallet_set_id", circleWalletSetID)
	}
	if circleWalletSetName := os.Getenv("CIRCLE_DEFAULT_WALLET_SET_NAME"); circleWalletSetName != "" {
		viper.Set("circle.default_wallet_set_name", circleWalletSetName)
	}
	if supportedChains := os.Getenv("CIRCLE_SUPPORTED_CHAINS"); supportedChains != "" {
		parts := strings.Split(supportedChains, ",")
		var chains []string
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				chains = append(chains, strings.ToUpper(trimmed))
			}
		}
		if len(chains) > 0 {
			viper.Set("circle.supported_chains", chains)
		}
	}
	if circleEnv := os.Getenv("CIRCLE_ENVIRONMENT"); circleEnv != "" {
		viper.Set("circle.environment", circleEnv)
	}

	// KYC Provider
	if kycAPIKey := os.Getenv("KYC_API_KEY"); kycAPIKey != "" {
		viper.Set("kyc.api_key", kycAPIKey)
	}
	if sumsubToken := os.Getenv("SUMSUB_APP_TOKEN"); sumsubToken != "" {
		viper.Set("kyc.api_key", sumsubToken)
		viper.Set("kyc.provider", "sumsub")
	}
	if kycAPISecret := os.Getenv("KYC_API_SECRET"); kycAPISecret != "" {
		viper.Set("kyc.api_secret", kycAPISecret)
	}
	if sumsubSecret := os.Getenv("SUMSUB_SECRET_KEY"); sumsubSecret != "" {
		viper.Set("kyc.api_secret", sumsubSecret)
	}
	if kycProvider := os.Getenv("KYC_PROVIDER"); kycProvider != "" {
		viper.Set("kyc.provider", kycProvider)
	}
	if kycCallbackURL := os.Getenv("KYC_CALLBACK_URL"); kycCallbackURL != "" {
		viper.Set("kyc.callback_url", kycCallbackURL)
	}
	if kycBaseURL := os.Getenv("KYC_BASE_URL"); kycBaseURL != "" {
		viper.Set("kyc.base_url", kycBaseURL)
	}
	if sumsubBaseURL := os.Getenv("SUMSUB_BASE_URL"); sumsubBaseURL != "" {
		viper.Set("kyc.base_url", sumsubBaseURL)
	}
	if kycLevelName := os.Getenv("KYC_LEVEL_NAME"); kycLevelName != "" {
		viper.Set("kyc.level_name", kycLevelName)
	}
	if sumsubLevelName := os.Getenv("SUMSUB_LEVEL_NAME"); sumsubLevelName != "" {
		viper.Set("kyc.level_name", sumsubLevelName)
	}

	// Email Service
	if emailAPIKey := os.Getenv("EMAIL_API_KEY"); emailAPIKey != "" {
		viper.Set("email.api_key", emailAPIKey)
	}
	if resendAPIKey := os.Getenv("RESEND_API_KEY"); resendAPIKey != "" {
		viper.Set("email.api_key", resendAPIKey)
		viper.Set("email.provider", "resend")
	}
	if emailProvider := os.Getenv("EMAIL_PROVIDER"); emailProvider != "" {
		viper.Set("email.provider", emailProvider)
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		viper.Set("email.base_url", baseURL)
	}
	if emailBaseURL := os.Getenv("EMAIL_BASE_URL"); emailBaseURL != "" {
		viper.Set("email.base_url", emailBaseURL)
	}
	if fromEmail := os.Getenv("EMAIL_FROM_EMAIL"); fromEmail != "" {
		viper.Set("email.from_email", fromEmail)
	}
	if resendFrom := os.Getenv("RESEND_FROM_EMAIL"); resendFrom != "" {
		viper.Set("email.from_email", resendFrom)
	}
	if fromName := os.Getenv("EMAIL_FROM_NAME"); fromName != "" {
		viper.Set("email.from_name", fromName)
	}
	if resendFromName := os.Getenv("RESEND_FROM_NAME"); resendFromName != "" {
		viper.Set("email.from_name", resendFromName)
	}
	if replyTo := os.Getenv("EMAIL_REPLY_TO"); replyTo != "" {
		viper.Set("email.reply_to", replyTo)
	}

	// 0G Network
	// Storage configuration
	if zeroGStorageRPC := os.Getenv("ZEROG_STORAGE_RPC_ENDPOINT"); zeroGStorageRPC != "" {
		viper.Set("zerog.storage.rpc_endpoint", zeroGStorageRPC)
	}
	if zeroGIndexerRPC := os.Getenv("ZEROG_STORAGE_INDEXER_RPC"); zeroGIndexerRPC != "" {
		viper.Set("zerog.storage.indexer_rpc", zeroGIndexerRPC)
	}
	if zeroGStorageKey := os.Getenv("ZEROG_STORAGE_PRIVATE_KEY"); zeroGStorageKey != "" {
		viper.Set("zerog.storage.private_key", zeroGStorageKey)
	}

	// Compute configuration
	if zeroGComputeBroker := os.Getenv("ZEROG_COMPUTE_BROKER_ENDPOINT"); zeroGComputeBroker != "" {
		viper.Set("zerog.compute.broker_endpoint", zeroGComputeBroker)
	}
	if zeroGComputeKey := os.Getenv("ZEROG_COMPUTE_PRIVATE_KEY"); zeroGComputeKey != "" {
		viper.Set("zerog.compute.private_key", zeroGComputeKey)
	}
	if zeroGProviderID := os.Getenv("ZEROG_COMPUTE_PROVIDER_ID"); zeroGProviderID != "" {
		viper.Set("zerog.compute.provider_id", zeroGProviderID)
	}
}

func validate(config *Config) error {
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if config.Security.EncryptionKey == "" {
		return fmt.Errorf("encryption key is required")
	}

	if config.Database.URL == "" && (config.Database.Host == "" || config.Database.Name == "") {
		return fmt.Errorf("database configuration is incomplete")
	}

	if strings.TrimSpace(config.Circle.EntitySecretCiphertext) == "" {
		return fmt.Errorf("circle entity secret ciphertext is required")
	}

	if len(config.Circle.SupportedChains) == 0 {
		return fmt.Errorf("circle supported chains configuration is required")
	}

	return nil
}
