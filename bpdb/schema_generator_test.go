package bpdb

import (
	"reflect"
	"testing"

	"github.com/twitchscience/scoop_protocol/scoop_protocol"
)

func TestApplyOperationAddColumns(t *testing.T) {
	base := AnnotatedSchema{
		EventName: "video_ad_request_error",
		Columns: []scoop_protocol.ColumnDefinition{
			{InboundName: "backend", OutboundName: "backend", Transformer: "varchar", ColumnCreationOptions: "(32)", SupportingColumns: ""},
			{InboundName: "content_mode", OutboundName: "content_mode", Transformer: "varchar", ColumnCreationOptions: "(32)", SupportingColumns: ""},
			{InboundName: "quality", OutboundName: "quality", Transformer: "varchar", ColumnCreationOptions: "(16)", SupportingColumns: ""},
		},
	}
	ops := []scoop_protocol.Operation{
		{"add", "minutes_logged", map[string]string{"inbound": "minutes_logged", "column_type": "bigint", "column_options": "", "supporting_columns": ""}},
		{"delete", "backend", map[string]string{}},
		{"add", "os", map[string]string{"inbound": "os", "column_type": "varchar", "column_options": "(16)", "supporting_columns": ""}},
		{"add", "id", map[string]string{"inbound": "id", "column_type": "idVarchar", "column_options": "(32)", "supporting_columns": "os"}},
	}
	expected := AnnotatedSchema{
		EventName: "video_ad_request_error",
		Columns: []scoop_protocol.ColumnDefinition{
			{InboundName: "content_mode", OutboundName: "content_mode", Transformer: "varchar", ColumnCreationOptions: "(32)", SupportingColumns: ""},
			{InboundName: "quality", OutboundName: "quality", Transformer: "varchar", ColumnCreationOptions: "(16)", SupportingColumns: ""},
			{InboundName: "minutes_logged", OutboundName: "minutes_logged", Transformer: "bigint", ColumnCreationOptions: "", SupportingColumns: ""},
			{InboundName: "os", OutboundName: "os", Transformer: "varchar", ColumnCreationOptions: "(16)", SupportingColumns: ""},
			{InboundName: "id", OutboundName: "id", Transformer: "idVarchar", ColumnCreationOptions: "(32)", SupportingColumns: "os"},
		},
	}
	err := ApplyOperations(&base, ops)
	if err != nil || !reflect.DeepEqual(expected, base) {
		t.Errorf("Results schema differs from expected:\n%v\nvs\n%v.", base, expected)
	}
}

func TestApplyOperationAddDupeColumns(t *testing.T) {
	base := AnnotatedSchema{
		EventName: "video_ad_request_error",
		Columns: []scoop_protocol.ColumnDefinition{
			{InboundName: "backend", OutboundName: "backend", Transformer: "varchar", ColumnCreationOptions: "(32)", SupportingColumns: ""},
		},
	}
	ops := []scoop_protocol.Operation{
		{"add", "minutes_logged", map[string]string{"inbound": "minutes_logged", "column_type": "bigint", "column_options": "", "supporting_columns": ""}},
		{"add", "backend", map[string]string{"inbound": "ip", "column_type": "varchar", "column_options": "(32)", "supporting_columns": ""}},
	}
	err := ApplyOperations(&base, ops)
	if err == nil {
		t.Error("Expected error on adding existing row.")
	}
}

func TestApplyOperationDeleteNonExistentColumns(t *testing.T) {
	base := AnnotatedSchema{
		EventName: "video_ad_request_error",
		Columns: []scoop_protocol.ColumnDefinition{
			{InboundName: "backend", OutboundName: "backend", Transformer: "varchar", ColumnCreationOptions: "(32)"},
		},
	}
	ops := []scoop_protocol.Operation{
		{"delete", "minutes_logged", map[string]string{}}, // delete non-existent column
	}
	err := ApplyOperations(&base, ops)
	if err == nil {
		t.Error("Expected error on adding existing row.")
	}
}
