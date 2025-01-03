package reeemiks

import (
	"fmt"
	"net"
	// "regexp"

	"github.com/jfreymuth/pulse/proto"
	"go.uber.org/zap"
)

type paSessionFinder struct {
	logger        *zap.SugaredLogger
	sessionLogger *zap.SugaredLogger
	config        *CanonicalConfig

	client *proto.Client
	conn   net.Conn
}

func newSessionFinder(logger *zap.SugaredLogger, config *CanonicalConfig) (SessionFinder, error) {
	client, conn, err := proto.Connect("")
	if err != nil {
		logger.Warnw("Failed to establish PulseAudio connection", "error", err)
		return nil, fmt.Errorf("establish PulseAudio connection: %w", err)
	}

	request := proto.SetClientName{
		Props: proto.PropList{
			"application.name": proto.PropListString("Reeemiks"),
		},
	}
	reply := proto.SetClientNameReply{}

	if err := client.Request(&request, &reply); err != nil {
		return nil, err
	}

	sf := &paSessionFinder{
		logger:        logger.Named("session_finder"),
		sessionLogger: logger.Named("sessions"),
		config:        config,
		client:        client,
		conn:          conn,
	}

	sf.logger.Debug("Created PA session finder instance")

	return sf, nil
}

func (sf *paSessionFinder) GetAllSessions() ([]Session, error) {
	sessions := []Session{}

	// get the master sink session
	masterSink, err := sf.getMasterSinkSession()
	if err == nil {
		sessions = append(sessions, masterSink)
	} else {
		sf.logger.Warnw("Failed to get master audio sink session", "error", err)
	}

	// get the master source session
	masterSource, err := sf.getMasterSourceSession()
	if err == nil {
		sessions = append(sessions, masterSource)
	} else {
		sf.logger.Warnw("Failed to get master audio source session", "error", err)
	}

	// enumerate sink inputs and add sessions along the way
	if err := sf.enumerateAndAddSessions(&sessions); err != nil {
		sf.logger.Warnw("Failed to enumerate audio sessions", "error", err)
		return nil, fmt.Errorf("enumerate audio sessions: %w", err)
	}

	return sessions, nil
}

func (sf *paSessionFinder) Release() error {
	if err := sf.conn.Close(); err != nil {
		sf.logger.Warnw("Failed to close PulseAudio connection", "error", err)
		return fmt.Errorf("close PulseAudio connection: %w", err)
	}

	sf.logger.Debug("Released PA session finder instance")

	return nil
}

func (sf *paSessionFinder) getMasterSinkSession() (Session, error) {
	request := proto.GetSinkInfo{
		SinkIndex: proto.Undefined,
	}
	reply := proto.GetSinkInfoReply{}

	if err := sf.client.Request(&request, &reply); err != nil {
		sf.logger.Warnw("Failed to get master sink info", "error", err)
		return nil, fmt.Errorf("get master sink info: %w", err)
	}

	// create the master sink session
	sink := newMasterSession(sf.sessionLogger, sf.client, reply.SinkIndex, reply.Channels, true)

	return sink, nil
}

func (sf *paSessionFinder) getMasterSourceSession() (Session, error) {
	request := proto.GetSourceInfo{
		SourceIndex: proto.Undefined,
	}
	reply := proto.GetSourceInfoReply{}

	if err := sf.client.Request(&request, &reply); err != nil {
		sf.logger.Warnw("Failed to get master source info", "error", err)
		return nil, fmt.Errorf("get master source info: %w", err)
	}

	// create the master source session
	source := newMasterSession(sf.sessionLogger, sf.client, reply.SourceIndex, reply.Channels, false)

	return source, nil
}

func (sf *paSessionFinder) enumerateAndAddSessions(sessions *[]Session) error {
	request := proto.GetSinkInputInfoList{}
	reply := proto.GetSinkInputInfoListReply{}

	if err := sf.client.Request(&request, &reply); err != nil {
		sf.logger.Warnw("Failed to get sink input list", "error", err)
		return fmt.Errorf("get sink input list: %w", err)
	}

	if sf.config.ReeemiksMatching == "default" {
		for _, info := range reply {
			name, ok := info.Properties["application.process.binary"]

			if !ok {
				sf.logger.Warnw("Failed to get sink input's process name", "sinkInputIndex", info.SinkInputIndex)
				continue
			}

			// create the reeemiks session object
			newSession := newPASession(sf.sessionLogger, sf.client, info.SinkInputIndex, info.Channels, name.String())

			// add it to our slice
			*sessions = append(*sessions, newSession)

		}
	} else { // All new Reeemiks features
		for _, info := range reply {
			var process_binary proto.PropListEntry
			var process_name proto.PropListEntry
			var name string
			var ok bool

			process_binary, ok = info.Properties["media.name"]
			if !ok {
				sf.logger.Debug("Failed to get sink input's name", "sinkInputIndex", info.SinkInputIndex)
				continue
			}
			process_name, ok = info.Properties["application.name"]
			if !ok {
				sf.logger.Debug("Failed to get sink input's application name", "sinkInputIndex", info.SinkInputIndex)
				continue
			}
			name = process_binary.String() + ": " + process_name.String()

			sf.logger.Info("Process: ", name)

			// create the reeemiks session object
			newSession := newPASession(sf.sessionLogger, sf.client, info.SinkInputIndex, info.Channels, name)

			// add it to our slice
			*sessions = append(*sessions, newSession)

		}

		request2 := proto.GetSinkInfoList{}
		reply2 := proto.GetSinkInfoListReply{}

		if err := sf.client.Request(&request2, &reply2); err != nil {
			sf.logger.Warnw("Failed to get sink list", "error", err)
			return fmt.Errorf("get sink list: %w", err)
		}

		for _, info := range reply2 {
			var audio_pos proto.PropListEntry
			var process_binary proto.PropListEntry
			var process_name proto.PropListEntry
			var name string
			var ok bool

			audio_pos, ok = info.Properties["audio.position"]
			if ok {
				audio_pos = audio_pos
				process_binary, ok = info.Properties["media.name"]
				if !ok {
					sf.logger.Debug("Failed to get sink's media name", "sinkInputIndex", info.SinkIndex)
					continue
				}
				process_name, ok = info.Properties["node.name"]
				if !ok {
					sf.logger.Debug("Failed to get sink's node name", "sinkInputIndex", info.SinkIndex)
					continue
				}
				// name = "reeemiks.device: " + process_name.String()
				name = "reeemiks.device: " + process_binary.String() + "~" + process_name.String()
			}

			// regex_match, err := regexp.MatchString("^(Audio/(Device|Sink|Source)|Stream/Output/Audio)$", media_class.String())
			// if err==nil {
			// 	if regex_match {
			// 	} else {
			// 	}
			// } else {
			// 	if !ok {
			// 		sf.logger.Warnw("Failed to regex match sink input's media class", "sinkInputIndex", info.SinkIndex)
			// 		continue
			// 	}
			// }

			if name != "" {
				sf.logger.Info("Sink: ", name)

				// create the reeemiks session object
				newSession := newPASession(sf.sessionLogger, sf.client, info.SinkIndex, info.Channels, name)

				// add it to our slice
				*sessions = append(*sessions, newSession)
			}
		}
	}

	return nil
}
