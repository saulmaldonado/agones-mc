package config

import (
	"time"

	"github.com/spf13/viper"
)

type Edition string
type Environment string
type Subcommand string

const (
	// subcommands

	Monitor Subcommand = "monitor"
	Backup  Subcommand = "backup"
	Load    Subcommand = "load"
)

const (
	// edition

	JavaEdition    Edition = "java"
	BedrockEdition Edition = "bedrock"

	// environment

	Development Environment = "development"
	Production  Environment = "production"
)

const (
	// shared config

	ENVIRONMENT   string = "ENVIRONMENT"
	INITIAL_DELAY string = "INITIAL_DELAY"

	// server config

	EDITION       string = "EDITION"
	PORT          string = "PORT"
	HOST          string = "HOST"
	RCON_PORT     string = "RCON_PORT"
	RCON_PASSWORD string = "RCON_PASSWORD"
	VOLUME        string = "VOLUME"
	POD_NAME      string = "POD_NAME"

	// monitor config

	MAX_ATTEMPTS string = "MAX_ATTEMPTS"
	INTERVAL     string = "INTERVAL"
	TIMEOUT      string = "TIMEOUT"

	// backup config

	BUCKET_NAME string = "BUCKET_NAME"
	BACKUP_CRON string = "BACKUP_CRON"
	BACKUP_NAME string = "BACKUP_NAME"
)

var (
	// defaults

	// shared

	ENVIRONMENT_DEFAULT   Environment   = Development
	INITIAL_DELAY_DEFAULT time.Duration = time.Second * 30

	// server config

	EDITION_DEFAULT       string = "java"
	PORT_DEFAULT          int    = 25565
	HOST_DEFAULT          string = "localhost"
	RCON_PORT_DEFAULT     int    = 25575
	RCON_PASSWORD_DEFAULT string = "minecraft"
	VOLUME_DEFAULT        string = "/data"
	POD_NAME_DEFAULT      string = ""

	// monitor config

	MAX_ATTEMPTS_DEFAULT int           = 5
	INTERVAL_DEFAULT     time.Duration = time.Second * 10
	TIMEOUT_DEFAULT      time.Duration = time.Second * 10

	// backup config

	BUCKET_NAME_DEFAULT string = ""
	BACKUP_CRON_DEFAULT string = ""
	BACKUP_NAME_DEFAULT string = ""
)

type SharedConfig interface {
	GetInitialDelay() time.Duration
	GetEnvironment() Environment
}

type ServerConfig interface {
	GetHost() string
	GetPort() int
	GetEdition() Edition
	GetRCONPort() int
	GetRCONPassword() string
	GetVolume() string
	GetPodName() string
}

type MonitorConfig interface {
	SharedConfig
	ServerConfig
	GetInterval() time.Duration
	GetTimeout() time.Duration
	GetAttempts() int
}

type BackupConfig interface {
	SharedConfig
	ServerConfig
	GetBucketName() string
	GetBackupCron() string
}

type LoadConfig interface {
	ServerConfig
	ServerConfig
	GetBucketName() string
	GetBackupName() string
}

type FileserverConfig interface {
	GetVolume() string
}

type sharedConfig struct{}

func NewSharedConfig() SharedConfig {
	return sharedConfig{}
}

func (sharedConfig) GetEnvironment() Environment {
	return Environment(viper.GetString(string(ENVIRONMENT)))
}

func (sharedConfig) GetInitialDelay() time.Duration {
	return viper.GetDuration(INITIAL_DELAY)
}

type serverConfig struct{}

func (serverConfig) GetHost() string {
	return viper.GetString(HOST)
}

func (serverConfig) GetPort() int {
	return viper.GetInt(PORT)
}

func (serverConfig) GetEdition() Edition {
	return Edition(viper.GetString(EDITION))
}

func (serverConfig) GetRCONPort() int {
	return viper.GetInt(RCON_PORT)
}

func (serverConfig) GetRCONPassword() string {
	return viper.GetString(RCON_PASSWORD)
}

func (serverConfig) GetVolume() string {
	return viper.GetString(VOLUME)
}

func (serverConfig) GetPodName() string {
	return viper.GetString(POD_NAME)
}

type monitorConfig struct {
	sharedConfig
	serverConfig
}

func NewMonitorConfig() monitorConfig {
	return monitorConfig{}
}

func (monitorConfig) GetInterval() time.Duration {
	return viper.GetDuration(INTERVAL)
}

func (monitorConfig) GetTimeout() time.Duration {
	return viper.GetDuration(TIMEOUT)
}

func (monitorConfig) GetAttempts() int {
	return viper.GetInt(MAX_ATTEMPTS)
}

type backupConfig struct {
	sharedConfig
	serverConfig
}

func NewBackupConfig() backupConfig {
	return backupConfig{}
}

func (backupConfig) GetBucketName() string {
	return viper.GetString(BUCKET_NAME)
}

func (backupConfig) GetBackupCron() string {
	return viper.GetString(BACKUP_CRON)
}

type loadConfig struct {
	sharedConfig
	serverConfig
}

func NewLoadConfig() loadConfig {
	return loadConfig{}
}

func (loadConfig) GetBucketName() string {
	return viper.GetString(BUCKET_NAME)
}

func (loadConfig) GetBackupName() string {
	return viper.GetString(BACKUP_NAME)
}

type fileServerConfig struct{}

func NewFileServerConfig() fileServerConfig {
	return fileServerConfig{}
}

func (fileServerConfig) GetVolume() string {
	return viper.GetString(VOLUME)
}

func init() {
	viper.SetDefault(INITIAL_DELAY, INITIAL_DELAY_DEFAULT)
	viper.SetDefault(HOST, HOST_DEFAULT)
	viper.SetDefault(PORT, PORT_DEFAULT)
	viper.SetDefault(EDITION, EDITION_DEFAULT)
	viper.SetDefault(RCON_PORT, RCON_PORT_DEFAULT)
	viper.SetDefault(RCON_PASSWORD, RCON_PASSWORD_DEFAULT)
	viper.SetDefault(VOLUME, VOLUME_DEFAULT)
	viper.SetDefault(POD_NAME, POD_NAME_DEFAULT)
	viper.SetDefault(INTERVAL, INTERVAL_DEFAULT)
	viper.SetDefault(TIMEOUT, TIMEOUT_DEFAULT)
	viper.SetDefault(MAX_ATTEMPTS, MAX_ATTEMPTS_DEFAULT)
	viper.SetDefault(BUCKET_NAME, BUCKET_NAME_DEFAULT)
	viper.SetDefault(BACKUP_CRON, BACKUP_CRON_DEFAULT)
	viper.SetDefault(BACKUP_NAME, BACKUP_NAME_DEFAULT)

	viper.AutomaticEnv()
}
