// Code generated by https://github.com/henrymbaldwin/solana-anchor-go. DO NOT EDIT.

package contract_bindings_solana

import (
	"encoding/base64"
	"fmt"
	ag_binary "github.com/gagliardetto/binary"
	"reflect"
	"strings"
)

var eventTypes = map[[8]byte]reflect.Type{}
var eventNames = map[[8]byte]string{}
var (
	_ *strings.Builder = nil
)
var (
	_ *base64.Encoding = nil
)
var (
	_ *ag_binary.Decoder = nil
)
var (
	_ *fmt.Formatter = nil
)

type Event struct {
	Name string
	Data EventData
}

type EventData interface {
	UnmarshalWithDecoder(decoder *ag_binary.Decoder) error
	isEventData()
}

const eventLogPrefix = "Program data: "

func DecodeEvents(logMessages []string) (evts []*Event, err error) {
	decoder := ag_binary.NewDecoderWithEncoding(nil, ag_binary.EncodingBorsh)
	for _, log := range logMessages {
		if strings.HasPrefix(log, eventLogPrefix) {
			eventBase64 := log[len(eventLogPrefix):]

			var eventBinary []byte
			if eventBinary, err = base64.StdEncoding.DecodeString(eventBase64); err != nil {
				err = fmt.Errorf("failed to decode event log: %s", eventBase64)
				return
			}

			eventDiscriminator := ag_binary.TypeID(eventBinary[:8])
			if eventType, ok := eventTypes[eventDiscriminator]; ok {
				eventData := reflect.New(eventType).Interface().(EventData)
				decoder.Reset(eventBinary)
				if err = eventData.UnmarshalWithDecoder(decoder); err != nil {
					err = fmt.Errorf("failed to unmarshal event %s: %w", eventType.String(), err)
					return
				}
				evts = append(evts, &Event{
					Name: eventNames[eventDiscriminator],
					Data: eventData,
				})
			}
		}
	}
	return
}
