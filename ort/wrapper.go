// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package ort

// #cgo pkg-config: libonnxruntime
// #include "ort_wrapper.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"unsafe"
)

const (
	DeviceCUDA = "Cuda"
	DeviceCPU  = "Cpu"
)

var (
	ortApiBase *C.OrtApiBase
	ortApi     *C.OrtApi
)

func init() {
	ortApiBase = C.OrtGetApiBase()
	ortApi = C.ort_get_api(ortApiBase)
}

type Env struct {
	inner *C.OrtEnv
}

type SessionOptions struct {
	inner *C.OrtSessionOptions
}

type Session struct {
	inner *C.OrtSession
}

type MemoryInfo struct {
	inner *C.OrtMemoryInfo
}

type Allocator struct {
	session *Session
	inner   *C.OrtAllocator
}

type RunOptions struct {
	inner *C.OrtRunOptions
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

type Value interface {
	Value() *C.OrtValue
	Close()
}

type valueBase struct {
	inner *C.OrtValue
}

func (t *valueBase) Close() {
	C.ort_release_value(ortApi, t.inner)
}

func (t *valueBase) Value() *C.OrtValue {
	return t.inner
}

type AllocatorType C.OrtAllocatorType

const (
	InvalidAllocator AllocatorType = iota - 1
	DeviceAllocator
	ArenaAllocator
)

type MemType C.OrtMemType

const (
	MemTypeCPUInput MemType = iota - 2
	MemTypeCPUOutput
	MemTypeDefault
	MemTypeCPU = MemTypeCPUOutput
)

type ONNXTensorElementDataType C.ONNXTensorElementDataType

const (
	ONNX_TENSOR_ELEMENT_DATA_TYPE_UNDEFINED ONNXTensorElementDataType = iota
	ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT
	ONNX_TENSOR_ELEMENT_DATA_TYPE_UINT8
	ONNX_TENSOR_ELEMENT_DATA_TYPE_INT8
	ONNX_TENSOR_ELEMENT_DATA_TYPE_UINT16
	ONNX_TENSOR_ELEMENT_DATA_TYPE_INT16
	ONNX_TENSOR_ELEMENT_DATA_TYPE_INT32
	ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64
	ONNX_TENSOR_ELEMENT_DATA_TYPE_STRING
	ONNX_TENSOR_ELEMENT_DATA_TYPE_BOOL
	ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT16
	ONNX_TENSOR_ELEMENT_DATA_TYPE_DOUBLE
	ONNX_TENSOR_ELEMENT_DATA_TYPE_UINT32
	ONNX_TENSOR_ELEMENT_DATA_TYPE_UINT64
	ONNX_TENSOR_ELEMENT_DATA_TYPE_COMPLEX64
	ONNX_TENSOR_ELEMENT_DATA_TYPE_COMPLEX128
	ONNX_TENSOR_ELEMENT_DATA_TYPE_BFLOAT16
	ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT8E4M3FN
	ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT8E4M3FNUZ
	ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT8E5M2
	ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT8E5M2FNUZ
	ONNX_TENSOR_ELEMENT_DATA_TYPE_UINT4
	ONNX_TENSOR_ELEMENT_DATA_TYPE_INT4
)

type OrtLoggingLevel C.OrtLoggingLevel

const (
	LoggingLevelVerbose OrtLoggingLevel = iota
	LoggingLevelInfo
	LoggingLevelWarning
	LoggingLevelError
	LoggingLevelFatal
)

func errFrom(status C.OrtStatusPtr) error {
	if status == nil {
		return nil
	}

	msg := C.ort_get_error_message(ortApi, status)
	str := C.GoString(msg)
	defer C.ort_release_status(ortApi, status)
	return errors.New(str)
}

func NewEnv(level OrtLoggingLevel, identifier string) (*Env, error) {
	id := C.CString(identifier)
	defer C.free(unsafe.Pointer(id))
	var out *C.OrtEnv

	if err := errFrom(C.ort_create_env(ortApi, C.OrtLoggingLevel(level), id, &out)); err != nil {
		return nil, err
	}

	return &Env{out}, nil
}

func (env *Env) Close() {
	C.ort_release_env(ortApi, env.inner)
}

func (env *Env) NewSession(modelPath string, options *SessionOptions) (*Session, error) {
	path := C.CString(modelPath)
	defer C.free(unsafe.Pointer(path))
	var out *C.OrtSession
	var opts *C.OrtSessionOptions
	if options != nil {
		opts = options.inner
	}

	if err := errFrom(C.ort_create_session(ortApi, env.inner, path, opts, &out)); err != nil {
		return nil, err
	}

	return &Session{out}, nil
}

func (s *Session) NewAllocator(memInfo *MemoryInfo) (*Allocator, error) {
	var mem_info *C.OrtMemoryInfo
	if memInfo != nil {
		mem_info = memInfo.inner
	}

	var out *C.OrtAllocator

	if err := errFrom(C.ort_create_allocator(ortApi, s.inner, mem_info, &out)); err != nil {
		return nil, err
	}

	return &Allocator{s, out}, nil
}

func (a *Allocator) InputName(index uint64) (string, error) {
	var s *C.char
	if err := errFrom(C.ort_session_get_input_name(ortApi, a.session.inner, C.size_t(index), a.inner, &s)); err != nil {
		return "", err
	}

	result := C.GoString(s)
	C.ort_allocator_free(a.inner, unsafe.Pointer(s))
	return result, nil
}

func (a *Allocator) Close() {
	C.ort_release_allocator(ortApi, a.inner)
}

func (s *Session) InputCount() (uint64, error) {
	var out C.size_t
	if err := errFrom(C.ort_session_get_input_count(ortApi, s.inner, &out)); err != nil {
		return 0, err
	}
	return uint64(out), nil
}

func (s *Session) Run(options *RunOptions, inputs map[string]Value, outputNames []string) ([]Value, error) {
	inputArr := []*C.OrtValue{}
	inputNames := []*C.char{}
	for k, v := range inputs {
		ptr := C.CString(k)
		inputNames = append(inputNames, ptr)
		defer C.free(unsafe.Pointer(ptr))
		if v != nil {
			inputArr = append(inputArr, v.Value())
		} else {
			inputArr = append(inputArr, nil)
		}
	}

	var runOpts *C.OrtRunOptions
	if options != nil {
		runOpts = options.inner
	}

	var inputNamesPtr **C.char
	if len(inputNames) > 0 {
		inputNamesPtr = &inputNames[0]
	}

	var inputsPtr **C.OrtValue
	if len(inputArr) > 0 {
		inputsPtr = &inputArr[0]
	}

	outNames := []*C.char{}
	for _, n := range outputNames {
		ptr := C.CString(n)
		outNames = append(outNames, ptr)
		defer C.free(unsafe.Pointer(ptr))
	}

	var outputNamesPtr **C.char
	if len(outNames) > 0 {
		outputNamesPtr = &outNames[0]
	}

	outputs := make([]*C.OrtValue, len(outNames))
	var outputsPtr **C.OrtValue
	if len(outNames) > 0 {
		outputsPtr = &outputs[0]
	}

	if err := errFrom(C.ort_run(ortApi, s.inner, runOpts, inputNamesPtr, inputsPtr, C.size_t(len(inputs)), outputNamesPtr, C.size_t(len(outputNames)), outputsPtr)); err != nil {
		return nil, err
	}

	result := make([]Value, len(outNames))
	for i, v := range outputs {
		result[i] = &valueBase{v}
	}

	return result, nil
}

func (s *Session) RunOnOutput(options *RunOptions, inputs map[string]Value, outputs map[string]Value) error {
	inputArr := []*C.OrtValue{}
	inputNames := []*C.char{}
	for k, v := range inputs {
		ptr := C.CString(k)
		inputNames = append(inputNames, ptr)
		defer C.free(unsafe.Pointer(ptr))
		if v != nil {
			inputArr = append(inputArr, v.Value())
		} else {
			inputArr = append(inputArr, nil)
		}
	}

	var runOpts *C.OrtRunOptions
	if options != nil {
		runOpts = options.inner
	}

	var inputNamesPtr **C.char
	if len(inputNames) > 0 {
		inputNamesPtr = &inputNames[0]
	}

	var inputsPtr **C.OrtValue
	if len(inputArr) > 0 {
		inputsPtr = &inputArr[0]
	}

	outputArr := []*C.OrtValue{}
	outputNames := []*C.char{}
	for k, v := range outputs {
		ptr := C.CString(k)
		outputNames = append(outputNames, ptr)
		defer C.free(unsafe.Pointer(ptr))
		if v != nil {
			outputArr = append(outputArr, v.Value())
		} else {
			outputArr = append(outputArr, nil)
		}
	}

	var outputNamesPtr **C.char
	if len(outputNames) > 0 {
		outputNamesPtr = &outputNames[0]
	}

	var outputsPtr **C.OrtValue
	if len(outputArr) > 0 {
		outputsPtr = &outputArr[0]
	}

	if err := errFrom(C.ort_run(ortApi, s.inner, runOpts, inputNamesPtr, inputsPtr, C.size_t(len(inputs)), outputNamesPtr, C.size_t(len(outputNames)), outputsPtr)); err != nil {
		return err
	}

	return nil
}

func (s *Session) Close() {
	C.ort_release_session(ortApi, s.inner)
}

func NewSessionOptions() (*SessionOptions, error) {
	var out *C.OrtSessionOptions
	if err := errFrom(C.ort_create_session_options(ortApi, &out)); err != nil {
		return nil, err
	}

	return &SessionOptions{out}, nil
}

func (so *SessionOptions) Close() {
	C.ort_release_session_options(ortApi, so.inner)
}

func NewMemoryInfo(name string, typ AllocatorType, id int, memTyp MemType) (*MemoryInfo, error) {
	var out *C.OrtMemoryInfo
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	if err := errFrom(C.ort_create_memory_info(ortApi, n, C.OrtAllocatorType(typ), C.int(id), C.OrtMemType(memTyp), &out)); err != nil {
		return nil, err
	}

	return &MemoryInfo{out}, nil
}

func (mi *MemoryInfo) NewTensorF32(data []float32, shape []int64) (*Tensor[float32], error) {
	var out *C.OrtValue
	var shapePtr *C.int64_t
	if len(shape) != 0 {
		shapePtr = (*C.int64_t)(&shape[0])
	}

	var dataPtr unsafe.Pointer
	if len(data) != 0 {
		dataPtr = unsafe.Pointer(&data[0])
	}

	if err := errFrom(C.ort_create_tensor_with_data_as_ort_value(ortApi, mi.inner, dataPtr, C.size_t(len(data))*4, shapePtr, C.size_t(len(shape)), C.ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT, &out)); err != nil {
		return nil, err
	}

	return &Tensor[float32]{valueBase{out}, data}, nil
}

type Tensor[T Number] struct {
	valueBase
	Data []T
}

func (t *Tensor[T]) Dims() ([]int64, error) {
	var typeAndShape *C.OrtTensorTypeAndShapeInfo
	if err := errFrom(C.ort_get_tensor_type_and_shape(ortApi, t.inner, &typeAndShape)); err != nil {
		return nil, err
	}

	defer C.ort_release_tensor_type_and_shape_info(ortApi, typeAndShape)

	var dimsCount C.size_t
	if err := errFrom(C.ort_get_dimensions_count(ortApi, typeAndShape, &dimsCount)); err != nil {
		return nil, err
	}

	out := make([]int64, dimsCount)
	if err := errFrom(C.ort_get_dimensions(ortApi, typeAndShape, (*C.int64_t)(&out[0]), dimsCount)); err != nil {
		return nil, err
	}
	return out, nil
}

func AsTensor[T Number](t Value) (*Tensor[T], error) {
	var out unsafe.Pointer
	if err := errFrom(C.ort_get_tensor_mutable_data(ortApi, t.Value(), &out)); err != nil {
		return nil, err
	}

	var sz uint64
	if err := errFrom(C.ort_get_tensor_size_in_bytes(ortApi, t.Value(), (*C.size_t)(&sz))); err != nil {
		return nil, err
	}

	var zero T
	n := sz / uint64(unsafe.Sizeof(zero))
	slice := unsafe.Slice((*T)(out), n)
	return &Tensor[T]{
		valueBase: valueBase{t.Value()},
		Data:      slice,
	}, nil
}

func (mi *MemoryInfo) Close() {
	C.ort_release_memory_info(ortApi, mi.inner)
}
