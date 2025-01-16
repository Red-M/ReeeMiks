// Package reeemiks provides a machine-side client that pairs with an Arduino
// chip to form a tactile, physical volume control system/
package reeemiks

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/Red-M/ReeeMiks/pkg/reeemiks/util"
)

const (

	// when this is set to anything, reeemiks won't use a tray icon
	envNoTray = "REEEMIKS_NO_TRAY_ICON"
)

// Reeemiks is the main entity managing access to all sub-components
type Reeemiks struct {
	logger         *zap.SugaredLogger
	notifier       Notifier
	config         *CanonicalConfig
	reeemiksConnection ReeemiksConnection
	sessions       *sessionMap

	stopChannel chan bool
	version     string
	verbose     bool
}

// NewReeemiks creates a Reeemiks instance
func NewReeemiks(logger *zap.SugaredLogger, verbose bool) (*Reeemiks, error) {
	logger = logger.Named("reeemiks")

	notifier, err := NewToastNotifier(logger)
	if err != nil {
		logger.Errorw("Failed to create ToastNotifier", "error", err)
		return nil, fmt.Errorf("create new ToastNotifier: %w", err)
	}

	config, err := NewConfig(logger, notifier)
	if err != nil {
		logger.Errorw("Failed to create Config", "error", err)
		return nil, fmt.Errorf("create new Config: %w", err)
	}

	d := &Reeemiks{
		logger:      logger,
		notifier:    notifier,
		config:      config,
		stopChannel: make(chan bool),
		verbose:     verbose,
	}

	sessionFinder, err := newSessionFinder(logger, config)
	if err != nil {
		logger.Errorw("Failed to create SessionFinder", "error", err)
		return nil, fmt.Errorf("create new SessionFinder: %w", err)
	}

	sessions, err := newSessionMap(d, logger, sessionFinder)
	if err != nil {
		logger.Errorw("Failed to create sessionMap", "error", err)
		return nil, fmt.Errorf("create new sessionMap: %w", err)
	}

	d.sessions = sessions

	logger.Debug("Created reeemiks instance")

	return d, nil
}

// Initialize sets up components and starts to run in the background
func (d *Reeemiks) Initialize() error {
	d.logger.Debug("Initializing")

	// load the config for the first time
	if err := d.config.Load(); err != nil {
		d.logger.Errorw("Failed to load config during initialization", "error", err)
		return fmt.Errorf("load config during init: %w", err)
	}

	if d.config.EnableHidListen {
		hid, err := NewHIDRAW(d, d.logger)
		if err != nil {
			d.logger.Errorw("Failed to create HIDRAW", "error", err)
			return fmt.Errorf("create new HIDRAW: %w", err)
		}

		// d.hid = hid
		d.reeemiksConnection = hid
	} else {
		serial, err := NewSerialIO(d, d.logger)
		if err != nil {
			d.logger.Errorw("Failed to create SerialIO", "error", err)
			return fmt.Errorf("create new SerialIO: %w", err)
		}

		// d.serial = serial
		d.reeemiksConnection = serial
	}

	// initialize the session map
	if err := d.sessions.initialize(); err != nil {
		d.logger.Errorw("Failed to initialize session map", "error", err)
		return fmt.Errorf("init session map: %w", err)
	}

	// decide whether to run with/without tray
	if _, noTraySet := os.LookupEnv(envNoTray); noTraySet {

		d.logger.Debugw("Running without tray icon", "reason", "envvar set")

		// run in main thread while waiting on ctrl+C
		d.setupInterruptHandler()
		d.run()

	} else {
		d.setupInterruptHandler()
		d.initializeTray(d.run)
	}

	return nil
}

// SetVersion causes reeemiks to add a version string to its tray menu if called before Initialize
func (d *Reeemiks) SetVersion(version string) {
	d.version = version
}

// Verbose returns a boolean indicating whether reeemiks is running in verbose mode
func (d *Reeemiks) Verbose() bool {
	return d.verbose
}

func (d *Reeemiks) setupInterruptHandler() {
	interruptChannel := util.SetupCloseHandler()

	go func() {
		signal := <-interruptChannel
		d.logger.Debugw("Interrupted", "signal", signal)
		d.signalStop()
	}()
}

func (d *Reeemiks) run() {
	d.logger.Info("Run loop starting")

	// watch the config file for changes
	go d.config.WatchConfigFileChanges()

	// connect to the arduino for the first time
	go func() {
		if err := d.reeemiksConnection.Start(); err != nil {
			d.logger.Warnw("Failed to start first-time serial connection", "error", err)

			// If the port is busy, that's because something else is connected - notify and quit
			if errors.Is(err, os.ErrPermission) {
				d.logger.Warnw("Serial port seems busy, notifying user and closing",
					"comPort", d.config.SerialConnectionInfo.COMPort)

				d.notifier.Notify(fmt.Sprintf("Can't connect to %s!", d.config.SerialConnectionInfo.COMPort),
					"This serial port is busy, make sure to close any serial monitor or other reeemiks instance.")

				d.signalStop()

				// also notify if the COM port they gave isn't found, maybe their config is wrong
			} else if errors.Is(err, os.ErrNotExist) {
				d.logger.Warnw("Provided COM port seems wrong, notifying user and closing",
					"comPort", d.config.SerialConnectionInfo.COMPort)

				d.notifier.Notify(fmt.Sprintf("Can't connect to %s!", d.config.SerialConnectionInfo.COMPort),
					"This serial port doesn't exist, check your configuration and make sure it's set correctly.")

				d.signalStop()
			}
		}
	}()

	// wait until stopped (gracefully)
	<-d.stopChannel
	d.logger.Debug("Stop channel signaled, terminating")

	if err := d.stop(); err != nil {
		d.logger.Warnw("Failed to stop reeemiks", "error", err)
		os.Exit(1)
	} else {
		// exit with 0
		os.Exit(0)
	}
}

func (d *Reeemiks) signalStop() {
	d.logger.Debug("Signalling stop channel")
	d.stopChannel <- true
}

func (d *Reeemiks) stop() error {
	d.logger.Info("Stopping")

	d.reeemiksConnection.Stop()
	d.config.StopWatchingConfigFile()

	// release the session map
	if err := d.sessions.release(); err != nil {
		d.logger.Errorw("Failed to release session map", "error", err)
		return fmt.Errorf("release session map: %w", err)
	}

	d.stopTray()

	// attempt to sync on exit - this won't necessarily work but can't harm
	d.logger.Sync()

	return nil
}
