// Copyright (c) 2013 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcwire_test

import (
	"bytes"
	"github.com/conformal/btcwire"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

// TestGetHeaders tests the MsgGetHeader API.
func TestGetHeaders(t *testing.T) {
	pver := btcwire.ProtocolVersion

	// Block 99500 hash.
	hashStr := "000000000002e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	locatorHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	// Ensure the command is expected value.
	wantCmd := "getheaders"
	msg := btcwire.NewMsgGetHeaders()
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgGetHeaders: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Protocol version 4 bytes + num hashes (varInt) + max block locator
	// hashes + hash stop.
	wantPayload := uint32(16045)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure block locator hashes are added properly.
	err = msg.AddBlockLocatorHash(locatorHash)
	if err != nil {
		t.Errorf("AddBlockLocatorHash: %v", err)
	}
	if msg.BlockLocatorHashes[0] != locatorHash {
		t.Errorf("AddBlockLocatorHash: wrong block locator added - "+
			"got %v, want %v",
			spew.Sprint(msg.BlockLocatorHashes[0]),
			spew.Sprint(locatorHash))
	}

	// Ensure adding more than the max allowed block locator hashes per
	// message returns an error.
	for i := 0; i < btcwire.MaxBlockLocatorsPerMsg; i++ {
		err = msg.AddBlockLocatorHash(locatorHash)
	}
	if err == nil {
		t.Errorf("AddBlockLocatorHash: expected error on too many " +
			"block locator hashes not received")
	}

	return
}

// TestGetHeadersWire tests the MsgGetHeaders wire encode and decode for various
// numbers of block locator hashes and protocol versions.
func TestGetHeadersWire(t *testing.T) {
	// Set protocol inside getheaders message.
	pver := uint32(60002)

	// Block 99499 hash.
	hashStr := "2710f40c87ec93d010a6fd95f42c59a2cbacc60b18cf6b7957535"
	hashLocator, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	// Block 99500 hash.
	hashStr = "2e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	hashLocator2, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	// Block 100000 hash.
	hashStr = "3ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	hashStop, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	// MsgGetHeaders message with no block locators or stop hash.
	NoLocators := btcwire.NewMsgGetHeaders()
	NoLocators.ProtocolVersion = pver
	NoLocatorsEncoded := []byte{
		0x62, 0xea, 0x00, 0x00, // Protocol version 60002
		0x00, // Varint for number of block locator hashes
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Hash stop
	}

	// MsgGetHeaders message with multiple block locators and a stop hash.
	MultiLocators := btcwire.NewMsgGetHeaders()
	MultiLocators.ProtocolVersion = pver
	MultiLocators.HashStop = *hashStop
	MultiLocators.AddBlockLocatorHash(hashLocator2)
	MultiLocators.AddBlockLocatorHash(hashLocator)
	MultiLocatorsEncoded := []byte{
		0x62, 0xea, 0x00, 0x00, // Protocol version 60002
		0x02, // Varint for number of block locator hashes
		0xe0, 0xde, 0x06, 0x44, 0x68, 0x13, 0x2c, 0x63,
		0xd2, 0x20, 0xcc, 0x69, 0x12, 0x83, 0xcb, 0x65,
		0xbc, 0xaa, 0xe4, 0x79, 0x94, 0xef, 0x9e, 0x7b,
		0xad, 0xe7, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // Block 99500 hash
		0x35, 0x75, 0x95, 0xb7, 0xf6, 0x8c, 0xb1, 0x60,
		0xcc, 0xba, 0x2c, 0x9a, 0xc5, 0x42, 0x5f, 0xd9,
		0x6f, 0x0a, 0x01, 0x3d, 0xc9, 0x7e, 0xc8, 0x40,
		0x0f, 0x71, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // Block 99499 hash
		0x06, 0xe5, 0x33, 0xfd, 0x1a, 0xda, 0x86, 0x39,
		0x1f, 0x3f, 0x6c, 0x34, 0x32, 0x04, 0xb0, 0xd2,
		0x78, 0xd4, 0xaa, 0xec, 0x1c, 0x0b, 0x20, 0xaa,
		0x27, 0xba, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, // Hash stop
	}

	tests := []struct {
		in   *btcwire.MsgGetHeaders // Message to encode
		out  *btcwire.MsgGetHeaders // Expected decoded message
		buf  []byte                 // Wire encoding
		pver uint32                 // Protocol version for wire encoding
	}{
		// Latest protocol version with no block locators.
		{
			NoLocators,
			NoLocators,
			NoLocatorsEncoded,
			btcwire.ProtocolVersion,
		},

		// Latest protocol version with multiple block locators.
		{
			MultiLocators,
			MultiLocators,
			MultiLocatorsEncoded,
			btcwire.ProtocolVersion,
		},

		// Protocol version BIP0035Version with no block locators.
		{
			NoLocators,
			NoLocators,
			NoLocatorsEncoded,
			btcwire.BIP0035Version,
		},

		// Protocol version BIP0035Version with multiple block locators.
		{
			MultiLocators,
			MultiLocators,
			MultiLocatorsEncoded,
			btcwire.BIP0035Version,
		},

		// Protocol version BIP0031Version with no block locators.
		{
			NoLocators,
			NoLocators,
			NoLocatorsEncoded,
			btcwire.BIP0031Version,
		},

		// Protocol version BIP0031Versionwith multiple block locators.
		{
			MultiLocators,
			MultiLocators,
			MultiLocatorsEncoded,
			btcwire.BIP0031Version,
		},

		// Protocol version NetAddressTimeVersion with no block locators.
		{
			NoLocators,
			NoLocators,
			NoLocatorsEncoded,
			btcwire.NetAddressTimeVersion,
		},

		// Protocol version NetAddressTimeVersion multiple block locators.
		{
			MultiLocators,
			MultiLocators,
			MultiLocatorsEncoded,
			btcwire.NetAddressTimeVersion,
		},

		// Protocol version MultipleAddressVersion with no block locators.
		{
			NoLocators,
			NoLocators,
			NoLocatorsEncoded,
			btcwire.MultipleAddressVersion,
		},

		// Protocol version MultipleAddressVersion multiple block locators.
		{
			MultiLocators,
			MultiLocators,
			MultiLocatorsEncoded,
			btcwire.MultipleAddressVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to wire format.
		var buf bytes.Buffer
		err := test.in.BtcEncode(&buf, test.pver)
		if err != nil {
			t.Errorf("BtcEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("BtcEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the message from wire format.
		var msg btcwire.MsgGetHeaders
		rbuf := bytes.NewBuffer(test.buf)
		err = msg.BtcDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("BtcDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&msg, test.out) {
			t.Errorf("BtcDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&msg), spew.Sdump(test.out))
			continue
		}
	}
}
