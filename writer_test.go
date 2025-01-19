package simpleini

import (
	"bytes"
	"errors"
	"testing"
)

type TestConfig struct {
	Field1       string  `ini:"field1"`
	Field2       *int    `ini:"field2"`
	Field3       bool    `ini:"field3"`
	Field4       float64 `ini:"field4"`
	Nested       *NestedConfig
	NonPtrNested NonPtrNestedConfig
	ExtraField   string `ini:"extra_field"`
}

type NestedConfig struct {
	Field5     string `ini:"field5"`
	Field6     *int   `ini:"field6"`
	SubNested  *SubNestedConfig
	ExtraField string `ini:"extra_field"`
}

type NonPtrNestedConfig struct {
	Field9          string `ini:"field9"`
	Field10         int    `ini:"field10"`
	SubNonPtrNested SubNonPtrNestedConfig
	ExtraField      string `ini:"extra_field"`
}

type SubNonPtrNestedConfig struct {
	Field11    bool   `ini:"field11"`
	Field12    string `ini:"field12"`
	ExtraField string `ini:"extra_field"`
}

type SubNestedConfig struct {
	Field7     *bool  `ini:"field7"`
	Field8     string `ini:"field8"`
	ExtraField string `ini:"extra_field"`
}

func TestWrite_Success(t *testing.T) {
	field2 := 42
	field6 := 100
	field7 := false

	config := &TestConfig{
		Field1: "value1",
		Field2: &field2,
		Field3: true,
		Field4: 3.14,
		Nested: &NestedConfig{
			Field5: "value5",
			Field6: &field6,
			SubNested: &SubNestedConfig{
				Field7: &field7,
				Field8: "value8",
			},
		},
		NonPtrNested: NonPtrNestedConfig{
			Field9:  "value9",
			Field10: 200,
			SubNonPtrNested: SubNonPtrNestedConfig{
				Field11: true,
				Field12: "value12",
			},
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `field1 = value1
field2 = 42
field3 = true
field4 = 3.14
extra_field = 

[nested]
field5 = value5
field6 = 100
extra_field = 

[nested.sub_nested]
field7 = false
field8 = value8
extra_field = 

[non_ptr_nested]
field9 = value9
field10 = 200
extra_field = 

[non_ptr_nested.sub_non_ptr_nested]
field11 = true
field12 = value12
extra_field = 
`
	if buf.String() != expected {
		t.Errorf("expected \n%q, got \n%q", expected, buf.String())
	}
}

func TestWrite_NilPointer(t *testing.T) {
	field2 := 42

	config := &TestConfig{
		Field1: "value1",
		Field2: &field2,
		Field3: true,
		Field4: 3.14,
		Nested: nil,
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `field1 = value1
field2 = 42
field3 = true
field4 = 3.14
extra_field = 

; [nested]
; field5 =
; field6 =
; extra_field =

; [nested.sub_nested]
; field7 =
; field8 =
; extra_field =

[non_ptr_nested]
field9 = 
field10 = 0
extra_field = 

[non_ptr_nested.sub_non_ptr_nested]
field11 = false
field12 = 
extra_field = 
`
	if buf.String() != expected {
		t.Errorf("expected %s, got %s", expected, buf.String())
	}
}

func TestWrite_InvalidConfig(t *testing.T) {
	config := "invalid config"

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedError := "configuration must be a pointer to a struct"
	if err.Error() != expectedError {
		t.Errorf("expected %s, got %s", expectedError, err.Error())
	}
}

func TestWrite_EmptyConfig(t *testing.T) {
	config := &TestConfig{}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `field1 = 
field2 = 
field3 = false
field4 = 0
extra_field = 

; [nested]
; field5 =
; field6 =
; extra_field =

; [nested.sub_nested]
; field7 =
; field8 =
; extra_field =

[non_ptr_nested]
field9 = 
field10 = 0
extra_field = 

[non_ptr_nested.sub_non_ptr_nested]
field11 = false
field12 = 
extra_field = 
`
	if buf.String() != expected {
		t.Errorf("expected %s, got %s", expected, buf.String())
	}
}

func TestWrite_MultipleSections(t *testing.T) {
	field2 := 42
	field6 := 100
	field7 := false

	config := &TestConfig{
		Field1: "value1",
		Field2: &field2,
		Field3: true,
		Field4: 3.14,
		Nested: &NestedConfig{
			Field5: "value5",
			Field6: &field6,
			SubNested: &SubNestedConfig{
				Field7:     &field7,
				Field8:     "value8",
				ExtraField: "extra_value8",
			},
			ExtraField: "extra_value6",
		},
		NonPtrNested: NonPtrNestedConfig{
			Field9:  "value9",
			Field10: 200,
			SubNonPtrNested: SubNonPtrNestedConfig{
				Field11:    true,
				Field12:    "value12",
				ExtraField: "extra_value12",
			},
			ExtraField: "extra_value10",
		},
		ExtraField: "extra_value1",
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `field1 = value1
field2 = 42
field3 = true
field4 = 3.14
extra_field = extra_value1

[nested]
field5 = value5
field6 = 100
extra_field = extra_value6

[nested.sub_nested]
field7 = false
field8 = value8
extra_field = extra_value8

[non_ptr_nested]
field9 = value9
field10 = 200
extra_field = extra_value10

[non_ptr_nested.sub_non_ptr_nested]
field11 = true
field12 = value12
extra_field = extra_value12
`
	if buf.String() != expected {
		t.Errorf("expected \n%s, got \n%s", expected, buf.String())
	}
}

func TestWrite_NilPointerSections(t *testing.T) {
	field2 := 42

	config := &TestConfig{
		Field1:     "value1",
		Field2:     &field2,
		Field3:     true,
		Field4:     3.14,
		Nested:     nil,
		ExtraField: "extra_value1",
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `field1 = value1
field2 = 42
field3 = true
field4 = 3.14
extra_field = extra_value1

; [nested]
; field5 =
; field6 =
; extra_field =

; [nested.sub_nested]
; field7 =
; field8 =
; extra_field =

[non_ptr_nested]
field9 = 
field10 = 0
extra_field = 

[non_ptr_nested.sub_non_ptr_nested]
field11 = false
field12 = 
extra_field = 
`
	if buf.String() != expected {
		t.Errorf("expected %s, got %s", expected, buf.String())
	}
}

func TestWrite_NilPointerField(t *testing.T) {
	config := &TestConfig{
		Field1: "value1",
		Field2: nil,
		Field3: true,
		Field4: 3.14,
		Nested: &NestedConfig{
			Field5: "value5",
			Field6: nil,
			SubNested: &SubNestedConfig{
				Field7: nil,
				Field8: "value8",
			},
		},
		NonPtrNested: NonPtrNestedConfig{
			Field9:  "value9",
			Field10: 200,
			SubNonPtrNested: SubNonPtrNestedConfig{
				Field11: true,
				Field12: "value12",
			},
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `field1 = value1
field2 = 
field3 = true
field4 = 3.14
extra_field = 

[nested]
field5 = value5
field6 = 
extra_field = 

[nested.sub_nested]
field7 = 
field8 = value8
extra_field = 

[non_ptr_nested]
field9 = value9
field10 = 200
extra_field = 

[non_ptr_nested.sub_non_ptr_nested]
field11 = true
field12 = value12
extra_field = 
`
	if buf.String() != expected {
		t.Errorf("expected \n%q, got \n%q", expected, buf.String())
	}
}

func TestWrite_InvalidFieldType(t *testing.T) {
	type InvalidConfig struct {
		Field1 complex128 `ini:"field1"`
	}

	config := &InvalidConfig{
		Field1: complex(1, 2),
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedError := "unsupported field type: complex128"
	if err.Error() != expectedError {
		t.Errorf("expected %s, got %s", expectedError, err.Error())
	}
}

func TestWrite_AnonymousStruct(t *testing.T) {
	type AnonymousConfig struct {
		TestConfig
		Field13 string `ini:"field13"`
	}

	field2 := 42
	field6 := 100
	field7 := false

	config := &AnonymousConfig{
		TestConfig: TestConfig{
			Field1: "value1",
			Field2: &field2,
			Field3: true,
			Field4: 3.14,
			Nested: &NestedConfig{
				Field5: "value5",
				Field6: &field6,
				SubNested: &SubNestedConfig{
					Field7: &field7,
					Field8: "value8",
				},
			},
			NonPtrNested: NonPtrNestedConfig{
				Field9:  "value9",
				Field10: 200,
				SubNonPtrNested: SubNonPtrNestedConfig{
					Field11: true,
					Field12: "value12",
				},
			},
		},
		Field13: "value13",
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `field1 = value1
field2 = 42
field3 = true
field4 = 3.14
extra_field = 
field13 = value13

[nested]
field5 = value5
field6 = 100
extra_field = 

[nested.sub_nested]
field7 = false
field8 = value8
extra_field = 

[non_ptr_nested]
field9 = value9
field10 = 200
extra_field = 

[non_ptr_nested.sub_non_ptr_nested]
field11 = true
field12 = value12
extra_field = 
`
	if buf.String() != expected {
		t.Errorf("expected \n%q, got \n%q", expected, buf.String())
	}
}

func TestWrite_AnonymousStructWithUnsupportedType(t *testing.T) {
	type InvalidStruct struct {
		Field14 complex128 `ini:"field14"`
	}

	type AnonymousConfigWithUnsupported struct {
		InvalidStruct
		TestConfig
	}

	field2 := 42

	config := &AnonymousConfigWithUnsupported{
		TestConfig: TestConfig{
			Field1: "value1",
			Field2: &field2,
			Field3: true,
			Field4: 3.14,
		},
		InvalidStruct: InvalidStruct{
			Field14: complex(1, 2),
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, config)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedError := "unsupported field type: complex128"
	if err.Error() != expectedError {
		t.Errorf("expected %s, got %s", expectedError, err.Error())
	}
}

// errorWriter is a custom io.Writer that always returns an error.
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("simulated write error")
}

func TestWrite_FprintfError(t *testing.T) {
	field2 := 42

	config := &TestConfig{
		Field1: "value1",
		Field2: &field2,
		Field3: true,
		Field4: 3.14,
	}

	w := &errorWriter{}
	err := Write(w, config)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedError := "simulated write error"
	if err.Error() != expectedError {
		t.Errorf("expected %s, got %s", expectedError, err.Error())
	}
}
