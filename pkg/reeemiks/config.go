package reeemiks

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/kirsle/configdir"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Red-M/ReeeMiks/pkg/reeemiks/util"
)

// CanonicalConfig provides application-wide access to configuration fields,
// as well as loading/file watching logic for reeemiks's configuration file
type CanonicalConfig struct {
	SliderMapping *sliderMap
	ButtonMapping map[string][]string

	SerialConnectionInfo struct {
		COMPort  string
		BaudRate int
	}

	HidConnectionInfo struct {
		VendorId  uint16
		ProductId uint16
		UsagePage uint16
		Usage     uint16
	}

	EnableHidListen bool
	InvertSliders bool
	NoiseReductionLevel string

	ReeemiksMatching string

	logger             *zap.SugaredLogger
	notifier           Notifier
	stopWatcherChannel chan bool

	reloadConsumers []chan bool

	userConfig     *viper.Viper
	internalConfig *viper.Viper
}

const (
	userConfigName     = "config"
	internalConfigName = "preferences"

	configType = "yaml"

	configKeySliderMapping       = "slider_mapping"
	configKeyButtonMapping       = "button_mapping"
	configKeyInvertSliders       = "invert_sliders"
	configKeyCOMPort             = "com_port"
	configKeyBaudRate            = "baud_rate"
	configKeyNoiseReductionLevel = "noise_reduction"
	configKeyVendorId            = "vendor_id"
	configKeyProductId           = "product_id"
	configKeyUsagePage           = "usage_page"
	configKeyUsage               = "usage"
	configKeyEnableHID           = "enable_hid_listen"
	configReeemiksMatching = "Reeemiks.matching"

	defaultCOMPort  = "COM4"
	defaultBaudRate = 9600
)

var userConfigFilename = userConfigName+"."+configType
var userConfigPath = func() string {
	configPath := configdir.LocalConfig("reeemiks")
	err := util.EnsureDirExists(configPath)
	if err != nil {
		panic(err)
	}
	// Check if we need to make XDG path
	// do existing config file check here
	// if XDG config doesn't exist and rel path does, move rel path to XDG
	// if XDG config does exist and rel config does exist, warn user
	// if XDG config doesn't exist and rel config doesn't exist, create XDG path

	configFile := path.Join(configPath, userConfigFilename)
	xdg_exists := util.FileExists(configFile)
	rel_exists := util.FileExists(userConfigFilename)

	if !xdg_exists && rel_exists {
		util.MoveFile(userConfigFilename,configFile)
	} else if xdg_exists && rel_exists {
		fmt.Printf("WARN: I'm ignoring your config relative to my binary, your config is located at: %s\n", configFile)
	} else if !xdg_exists && !rel_exists {
		fmt.Errorf("Config file doesn't exist: %s", configFile)
	}


	return configPath
}()
var userConfigFilepath = path.Join(userConfigPath, userConfigFilename)
var internalConfigPath = path.Join(userConfigPath, logDirectory)
var internalConfigFilepath = path.Join(userConfigPath, "preferences.yaml")

var defaultSliderMapping = func() *sliderMap {
	emptyMap := newSliderMap()
	emptyMap.set(0, []string{masterSessionName})

	return emptyMap
}()

// NewConfig creates a config instance for the reeemiks object and sets up viper instances for reeemiks's config files
func NewConfig(logger *zap.SugaredLogger, notifier Notifier) (*CanonicalConfig, error) {
	logger = logger.Named("config")

	cc := &CanonicalConfig{
		logger:             logger,
		notifier:           notifier,
		reloadConsumers:    []chan bool{},
		stopWatcherChannel: make(chan bool),
	}

	// distinguish between the user-provided config (config.yaml) and the internal config (logs/preferences.yaml)
	userConfig := viper.New()
	userConfig.SetConfigName(userConfigFilename)
	userConfig.SetConfigType(configType)
	userConfig.AddConfigPath(userConfigPath)

	userConfig.SetDefault(configKeySliderMapping, map[string][]string{})
	userConfig.SetDefault(configKeyButtonMapping, map[string][]string{})
	userConfig.SetDefault(configKeyInvertSliders, false)
	userConfig.SetDefault(configKeyCOMPort, defaultCOMPort)
	userConfig.SetDefault(configKeyBaudRate, defaultBaudRate)
	userConfig.SetDefault(configKeyEnableHID, false)
	userConfig.SetDefault(configReeemiksMatching, map[string]string{})

	internalConfig := viper.New()
	internalConfig.SetConfigName(internalConfigName)
	internalConfig.SetConfigType(configType)
	internalConfig.AddConfigPath(internalConfigPath)

	cc.userConfig = userConfig
	cc.internalConfig = internalConfig

	logger.Debug("Created config instance")

	return cc, nil
}

// Load reads reeemiks's config files from disk and tries to parse them
func (cc *CanonicalConfig) Load() error {
	cc.logger.Debugw("Loading config", "path", userConfigFilepath)

	// make sure it exists
	if !util.FileExists(userConfigFilepath) {
		cc.logger.Warnw("Config file not found", "path", userConfigFilepath)
		cc.notifier.Notify("Can't find configuration!",
			fmt.Sprintf("Config must be located at %s . Please re-launch", userConfigFilepath))

		return fmt.Errorf("Config file doesn't exist: %s", userConfigFilepath)
	}

	// load the user config
	if err := cc.userConfig.ReadInConfig(); err != nil {
		cc.logger.Warnw("Viper failed to read user config", "error", err)

		// if the error is yaml-format-related, show a sensible error. otherwise, show 'em to the logs
		if strings.Contains(err.Error(), "yaml:") {
			cc.notifier.Notify("Invalid configuration!",
				fmt.Sprintf("Please make sure %s is in a valid YAML format.", userConfigFilepath))
		} else {
			cc.notifier.Notify("Error loading configuration!", "Please check reeemiks's logs for more details.")
		}

		return fmt.Errorf("read user config: %w", err)
	}

	// load the internal config - this doesn't have to exist, so it can error
	if err := cc.internalConfig.ReadInConfig(); err != nil {
		cc.logger.Debugw("Viper failed to read internal config", "error", err, "reminder", "this is fine")
	}

	// canonize the configuration with viper's helpers
	if err := cc.populateFromVipers(); err != nil {
		cc.logger.Warnw("Failed to populate config fields", "error", err)
		return fmt.Errorf("populate config fields: %w", err)
	}

	cc.logger.Info("Loaded config successfully")
	cc.logger.Infow("Config values",
		"sliderMapping", cc.SliderMapping,
		"serialSonnectionInfo", cc.SerialConnectionInfo,
		"hidConectionInfo", cc.HidConnectionInfo,
		"invertSliders", cc.InvertSliders)

	return nil
}

// SubscribeToChanges allows external components to receive updates when the config is reloaded
func (cc *CanonicalConfig) SubscribeToChanges() chan bool {
	c := make(chan bool)
	cc.reloadConsumers = append(cc.reloadConsumers, c)

	return c
}

// WatchConfigFileChanges starts watching for configuration file changes
// and attempts reloading the config when they happen
func (cc *CanonicalConfig) WatchConfigFileChanges() {
	cc.logger.Debugw("Starting to watch user config file for changes", "path", userConfigFilepath)

	const (
		minTimeBetweenReloadAttempts = time.Millisecond * 500
		delayBetweenEventAndReload   = time.Millisecond * 50
	)

	lastAttemptedReload := time.Now()

	// establish watch using viper as opposed to doing it ourselves, though our internal cooldown is still required
	cc.userConfig.WatchConfig()
	cc.userConfig.OnConfigChange(func(event fsnotify.Event) {

		// when we get a write event...
		if event.Op&fsnotify.Write == fsnotify.Write {

			now := time.Now()

			// ... check if it's not a duplicate (many editors will write to a file twice)
			if lastAttemptedReload.Add(minTimeBetweenReloadAttempts).Before(now) {

				// and attempt reload if appropriate
				cc.logger.Debugw("Config file modified, attempting reload", "event", event)

				// wait a bit to let the editor actually flush the new file contents to disk
				<-time.After(delayBetweenEventAndReload)

				if err := cc.Load(); err != nil {
					cc.logger.Warnw("Failed to reload config file", "error", err)
				} else {
					cc.logger.Info("Reloaded config successfully")
					cc.notifier.Notify("Configuration reloaded!", "Your changes have been applied.")

					cc.onConfigReloaded()
				}

				// don't forget to update the time
				lastAttemptedReload = now
			}
		}
	})

	// wait till they stop us
	<-cc.stopWatcherChannel
	cc.logger.Debug("Stopping user config file watcher")
	cc.userConfig.OnConfigChange(nil)
}

// StopWatchingConfigFile signals our filesystem watcher to stop
func (cc *CanonicalConfig) StopWatchingConfigFile() {
	cc.stopWatcherChannel <- true
}

func (cc *CanonicalConfig) populateFromVipers() error {

	// merge the slider mappings from the user and internal configs
	cc.SliderMapping = sliderMapFromConfigs(
		cc.userConfig.GetStringMapStringSlice(configKeySliderMapping),
		cc.internalConfig.GetStringMapStringSlice(configKeySliderMapping),
	)

	// Get HID Config
	cc.EnableHidListen = cc.userConfig.GetBool(configKeyEnableHID)

	cc.HidConnectionInfo.ProductId = uint16(cc.userConfig.GetUint32(configKeyProductId))
	cc.HidConnectionInfo.VendorId = uint16(cc.userConfig.GetUint32(configKeyVendorId))
	cc.HidConnectionInfo.UsagePage = uint16(cc.userConfig.GetUint32(configKeyUsagePage))
	cc.HidConnectionInfo.Usage = uint16(cc.userConfig.GetUint32(configKeyUsage))

	// get the rest of the config fields - viper saves us a lot of effort here
	cc.SerialConnectionInfo.COMPort = cc.userConfig.GetString(configKeyCOMPort)

	cc.SerialConnectionInfo.BaudRate = cc.userConfig.GetInt(configKeyBaudRate)
	if cc.SerialConnectionInfo.BaudRate <= 0 && cc.EnableHidListen == false {
		cc.logger.Warnw("Invalid baud rate specified, using default value",
			"key", configKeyBaudRate,
			"invalidValue", cc.SerialConnectionInfo.BaudRate,
			"defaultValue", defaultBaudRate)

		cc.SerialConnectionInfo.BaudRate = defaultBaudRate
	}

	cc.InvertSliders = cc.userConfig.GetBool(configKeyInvertSliders)
	cc.NoiseReductionLevel = cc.userConfig.GetString(configKeyNoiseReductionLevel)

	cc.ReeemiksMatching = cc.userConfig.GetString(configReeemiksMatching)

	cc.logger.Debug("Populated config fields from vipers")

	return nil
}

func (cc *CanonicalConfig) onConfigReloaded() {
	cc.logger.Debug("Notifying consumers about configuration reload")

	for _, consumer := range cc.reloadConsumers {
		consumer <- true
	}
}
