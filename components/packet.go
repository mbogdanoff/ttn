// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"encoding/base64"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	"github.com/thethingsnetwork/core/semtech"
	"github.com/thethingsnetwork/core/utils/pointer"
	"reflect"
	"strings"
)

// ConvertRXPK create a core.Packet from a semtech.RXPK. It's an handy way to both decode the
// frame payload and retrieve associated metadata from that packet
func ConvertRXPK(p semtech.RXPK) (core.Packet, error) {
	packet := core.Packet{}
	if p.Data == nil {
		return packet, ErrImpossibleConversion
	}

	encoded := *p.Data
	switch len(encoded) % 4 {
	case 2:
		encoded += "=="
	case 3:
		encoded += "="
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return packet, err
	}

	payload := lorawan.NewPHYPayload(true)
	if err = payload.UnmarshalBinary(raw); err != nil {
		return packet, err
	}

	metadata := Metadata{}
	rxpkValue := reflect.ValueOf(p)
	rxpkStruct := rxpkValue.Type()
	metas := reflect.ValueOf(&metadata).Elem()
	for i := 0; i < rxpkStruct.NumField(); i += 1 {
		field := rxpkStruct.Field(i).Name
		if metas.FieldByName(field).CanSet() {
			metas.FieldByName(field).Set(rxpkValue.Field(i))
		}
	}

	return core.Packet{Metadata: &metadata, Payload: payload}, nil
}

// ConvertToTXPK converts a core Packet to a semtech TXPK packet using compatible metadata.
func ConvertToTXPK(p core.Packet) (semtech.TXPK, error) {
	raw, err := p.Payload.MarshalBinary()
	if err != nil {
		return semtech.TXPK{}, ErrImpossibleConversion
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")

	txpk := semtech.TXPK{Data: pointer.String(data)}
	if p.Metadata == nil {
		return txpk, nil
	}

	metadataValue := reflect.ValueOf(p.Metadata).Elem()
	metadataStruct := metadataValue.Type()
	txpkStruct := reflect.ValueOf(&txpk).Elem()
	for i := 0; i < metadataStruct.NumField(); i += 1 {
		field := metadataStruct.Field(i).Name
		if txpkStruct.FieldByName(field).CanSet() {
			txpkStruct.FieldByName(field).Set(metadataValue.Field(i))
		}
	}

	return txpk, nil
}
